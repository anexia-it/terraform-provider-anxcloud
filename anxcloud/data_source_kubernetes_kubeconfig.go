package anxcloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"go.anx.io/go-anxcloud/pkg/api"
)

func dataSourceKubernetesKubeconfig() *schema.Resource {
	kubeconfigSchema := schemaKubernetesKubeConfig()
	kubeconfigSchema["cluster"].ForceNew = false
	kubeconfigSchema["api_environment"].ForceNew = false

	return &schema.Resource{
		Description: "Retrieves an existing Kubernetes kubeconfig without requesting, regenerating, or removing it.",
		ReadContext: dataSourceKubernetesKubeconfigRead,
		Schema:      kubeconfigSchema,
	}
}

func dataSourceKubernetesKubeconfigRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	clusterID := d.Get("cluster").(string)
	cluster := kubernetesKubeconfigCluster{
		Identifier:     clusterID,
		apiEnvironment: d.Get("api_environment").(string),
	}
	if err := apiFromProviderConfig(m).Get(ctx, &cluster); err != nil {
		if api.IgnoreNotFound(err) == nil {
			return diag.Errorf("Kubernetes cluster %q was not found", clusterID)
		}
		return diag.Errorf("failed retrieving kubeconfig for Kubernetes cluster %q: %s", clusterID, err)
	}
	if cluster.KubeConfig == nil || *cluster.KubeConfig == "" {
		return diag.Errorf("Kubernetes cluster %q does not have an existing kubeconfig; create one separately before using this data source", clusterID)
	}

	diags := setResourceDataFromKubernetesKubeconfig(d, *cluster.KubeConfig)
	if !diags.HasError() {
		d.SetId(clusterID)
	}
	return diags
}
