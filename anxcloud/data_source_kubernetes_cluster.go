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
				"version",
				"location",
				"needs_service_vms",
				"enable_nat_gateways",
				"enable_lbaas",
				"internal_ipv4_prefix",
				"external_ipv4_prefix",
				"external_ipv6_prefix",
				"enable_autoscaling",
				"apiserver_allowlist",
				"cni_plugin",
				"external_ip_families",
				"enable_oidc_authentication",
				"oidc_client_id",
				"oidc_issuer_url",
				"oidc_groups_claim",
				"oidc_username_claim",
				"oidc_extra_scopes",
				"oidc_groups_prefix",
				"oidc_required_claim",
				"oidc_username_prefix",
				"maintenance_window_start_time",
				"maintenance_window_duration",
				"patch_version",
				"state",
				"state_text",
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
	identifier := d.Get("id").(string)
	if identifier == "" {
		foundCluster, err := findClusterByName(ctx, apiFromProviderConfig(m), d.Get("name").(string))
		if err != nil {
			return diag.Errorf("failed retrieving cluster by name: %s", err)
		}
		identifier = foundCluster.Identifier
	}

	cluster, err := getKubernetesCluster(ctx, apiFromProviderConfig(m), identifier)
	if err != nil {
		return diag.Errorf("failed retrieving cluster by id: %s", err)
	}

	d.SetId(cluster.Identifier)
	return setResourceDataFromKubernetesClusterV2(d, cluster)
}
