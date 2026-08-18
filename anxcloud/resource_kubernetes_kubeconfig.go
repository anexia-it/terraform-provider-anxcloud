package anxcloud

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"go.anx.io/go-anxcloud/pkg/api"
	"k8s.io/client-go/tools/clientcmd"
)

const kubernetesKubeconfigPollInterval = 5 * time.Second

func resourceKubernetesKubeconfig() *schema.Resource {
	return &schema.Resource{
		Description: strings.TrimSpace(`
			Resource to create a Kubernetes kubeconfig.
		`),

		CreateContext: resourceKubernetesKubeconfigCreate,
		ReadContext:   resourceKubernetesKubeconfigRead,
		DeleteContext: resourceKubernetesKubeconfigDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Read:   schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: schemaKubernetesKubeConfig(),
	}
}

func resourceKubernetesKubeconfigCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	d.SetId(d.Get("cluster").(string))
	return resourceKubernetesKubeconfigRead(ctx, d, m)
}

func resourceKubernetesKubeconfigRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	rawKubeconfig, err := getKubernetesKubeconfig(
		ctx,
		apiFromProviderConfig(m),
		d.Id(),
		d.Get("api_environment").(string),
	)
	if err != nil {
		return diag.Errorf("failed requesting kubeconfig: %s", err)
	}

	return setResourceDataFromKubernetesKubeconfig(d, rawKubeconfig)
}

func setResourceDataFromKubernetesKubeconfig(d *schema.ResourceData, rawKubeconfig string) diag.Diagnostics {

	kubeconfig, err := clientcmd.Load([]byte(rawKubeconfig))
	if err != nil {
		return diag.Errorf("failed deserializing kubeconfig: %s", err)
	}

	kubecontext := kubeconfig.Contexts[kubeconfig.CurrentContext]
	if kubecontext == nil {
		return diag.Errorf("kubeconfig current context %q was not found", kubeconfig.CurrentContext)
	}
	authInfo := kubeconfig.AuthInfos[kubecontext.AuthInfo]
	if authInfo == nil {
		return diag.Errorf("kubeconfig auth info %q was not found", kubecontext.AuthInfo)
	}
	cluster := kubeconfig.Clusters[kubecontext.Cluster]
	if cluster == nil {
		return diag.Errorf("kubeconfig cluster %q was not found", kubecontext.Cluster)
	}

	var diags diag.Diagnostics
	if err := d.Set("host", cluster.Server); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("cluster_ca_certificate", string(cluster.CertificateAuthorityData)); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("token", authInfo.Token); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("raw", rawKubeconfig); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	return diags
}

func resourceKubernetesKubeconfigDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	request := kubernetesKubeconfigRequest{
		Cluster:        d.Id(),
		apiEnvironment: d.Get("api_environment").(string),
	}
	if err := apiFromProviderConfig(m).Destroy(ctx, &request); api.IgnoreNotFound(err) != nil {
		return diag.Errorf("failed deleting kubeconfig: %s", err)
	}

	return nil
}

func getKubernetesKubeconfig(ctx context.Context, a api.API, clusterID, apiEnvironment string) (string, error) {
	ticker := time.NewTicker(kubernetesKubeconfigPollInterval)
	defer ticker.Stop()

	kubeconfigRequested := false
	for {
		cluster := kubernetesKubeconfigCluster{
			Identifier:     clusterID,
			apiEnvironment: apiEnvironment,
		}
		if err := a.Get(ctx, &cluster); err != nil {
			return "", fmt.Errorf("failed to get cluster: %w", err)
		}
		if cluster.KubeConfig != nil && *cluster.KubeConfig != "" {
			return *cluster.KubeConfig, nil
		}

		if !kubeconfigRequested {
			request := kubernetesKubeconfigRequest{
				Cluster:        clusterID,
				apiEnvironment: apiEnvironment,
			}
			if err := a.Create(ctx, &request); err != nil {
				return "", fmt.Errorf("failed to request kubeconfig: %w", err)
			}
			kubeconfigRequested = true
		}

		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-ticker.C:
		}
	}
}
