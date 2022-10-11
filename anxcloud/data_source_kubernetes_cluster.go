package anxcloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"go.anx.io/go-anxcloud/pkg/api"
	"go.anx.io/go-anxcloud/pkg/api/types"
	kubernetesv1 "go.anx.io/go-anxcloud/pkg/apis/kubernetes/v1"
)

func dataSourceKubernetesCluster() *schema.Resource {
	return &schema.Resource{
		Description: "Retrieves a Kubernetes cluster resource.",
		ReadContext: dataSourceKubernetesClusterRead,
		Schema: schemaWith(schemaKubernetesCluster(),
			fieldsExactlyOneOf("id", "name"),
			fieldsComputed(
				"location",
				"needs_service_vms",
				"enable_nat_gateways",
				"enable_lbaas",
			),
		),
	}
}

func findClusterByName(ctx context.Context, a api.API, name string) (*kubernetesv1.Cluster, error) {
	var channel types.ObjectChannel
	if err := a.List(ctx, &kubernetesv1.Cluster{}, api.ObjectChannel(&channel)); err != nil {
		return nil, fmt.Errorf("failed listing clusters: %s", err)
	}

	var listResult kubernetesv1.Cluster
	for retriever := range channel {
		if err := retriever(&listResult); err != nil {
			return nil, fmt.Errorf("failed retrieving cluster: %s", err)
		}

		if listResult.Name == name {
			if err := a.Get(ctx, &listResult); err != nil {
				return nil, fmt.Errorf("failed retrieving full cluster object: %w", err)
			}

			return &listResult, nil
		}
	}

	return nil, api.ErrNotFound
}

func dataSourceKubernetesClusterRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	a := apiFromProviderConfig(m)

	cluster := kubernetesv1.Cluster{
		Identifier: d.Get("id").(string),
		Name:       d.Get("name").(string),
	}

	if cluster.Identifier == "" {
		foundCluster, err := findClusterByName(ctx, a, cluster.Name)
		if err != nil {
			return diag.Errorf("failed retrieving cluster by name: %s", err)
		}
		cluster = *foundCluster
	} else {
		if err := a.Get(ctx, &cluster); err != nil {
			return diag.Errorf("failed retrieving cluster by id: %s", err)
		}
	}

	d.SetId(cluster.Identifier)

	return setResourceDataFromKubernetesCluster(ctx, a, d, cluster)
}
