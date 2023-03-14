package anxcloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	kubernetesv1 "go.anx.io/go-anxcloud/pkg/apis/kubernetes/v1"
	"go.anx.io/go-anxcloud/pkg/utils/pointer"
)

const gibiFactor = 1073741824 // math.Pow(2, 30)

func setResourceDataFromKubernetesCluster(d *schema.ResourceData, cluster kubernetesv1.Cluster) diag.Diagnostics {
	var diags diag.Diagnostics

	if err := d.Set("name", cluster.Name); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("location", cluster.Location.Identifier); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("version", cluster.Version); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("needs_service_vms", pointer.BoolVal(cluster.NeedsServiceVMs)); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("enable_nat_gateways", pointer.BoolVal(cluster.EnableNATGateways)); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("enable_lbaas", pointer.BoolVal(cluster.EnableLBaaS)); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if cluster.InternalIPv4Prefix != nil {
		if err := d.Set("internal_ipv4_prefix", cluster.InternalIPv4Prefix.Identifier); err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
	}
	if cluster.ExternalIPv4Prefix != nil {
		if err := d.Set("external_ipv4_prefix", cluster.ExternalIPv4Prefix.Identifier); err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
	}
	if cluster.ExternalIPv6Prefix != nil {
		if err := d.Set("external_ipv6_prefix", cluster.ExternalIPv6Prefix.Identifier); err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
	}

	return diags
}
