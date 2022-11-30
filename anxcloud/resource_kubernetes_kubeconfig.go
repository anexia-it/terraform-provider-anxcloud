package anxcloud

import (
	"context"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"go.anx.io/go-anxcloud/pkg/api"
	kubernetesv1 "go.anx.io/go-anxcloud/pkg/apis/kubernetes/v1"
	"k8s.io/client-go/tools/clientcmd"
)

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
	rawKubeconfig, err := kubernetesv1.GetKubeConfig(ctx, apiFromProviderConfig(m), d.Id())
	if err != nil {
		return diag.Errorf("failed requesting kubeconfig: %s", err)
	}

	kubeconfig, err := clientcmd.Load([]byte(rawKubeconfig))
	if err != nil {
		return diag.Errorf("failed loading deserializing kubeconfig: %s", err)
	}

	kubecontext := kubeconfig.Contexts[kubeconfig.CurrentContext]
	authInfo := kubeconfig.AuthInfos[kubecontext.AuthInfo]
	cluster := kubeconfig.Clusters[kubecontext.Cluster]

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
	if err := kubernetesv1.RemoveKubeConfig(ctx, apiFromProviderConfig(m), d.Id()); api.IgnoreNotFound(err) != nil {
		return diag.Errorf("failed deleting kubeconfig: %s", err)
	}

	return nil
}
