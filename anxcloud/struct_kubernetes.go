package anxcloud

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const gibiFactor = 1073741824 // math.Pow(2, 30)

func setResourceDataFromKubernetesClusterV2(d *schema.ResourceData, cluster kubernetesCluster) diag.Diagnostics {
	var diags diag.Diagnostics
	set := func(name string, value any) {
		if err := d.Set(name, value); err != nil {
			diags = append(diags, diag.FromErr(fmt.Errorf("set %s: %w", name, err))...)
		}
	}

	set("name", cluster.Name)
	set("location", cluster.Location.Identifier)
	set("version", cluster.Version.ID)
	set("patch_version", cluster.PatchVersion)
	set("needs_service_vms", cluster.NeedsServiceVMs)
	set("enable_nat_gateways", cluster.EnableNAT)
	set("enable_lbaas", cluster.EnableLBaaS)
	set("enable_autoscaling", cluster.EnableAutoscaling)
	set("cni_plugin", cluster.CNIPlugin.ID)
	set("enable_persistent_storage", cluster.EnablePersistentStorage)
	set("external_ip_families", cluster.ExternalIPFamilies.ID)
	set("enable_oidc_authentication", cluster.EnableOIDCAuthentication)
	set("oidc_client_id", cluster.OIDCClientID)
	set("oidc_issuer_url", cluster.OIDCIssuerURL)
	set("oidc_groups_claim", cluster.OIDCGroupsClaim)
	set("oidc_username_claim", cluster.OIDCUsernameClaim)
	set("oidc_extra_scopes", cluster.OIDCExtraScopes)
	set("oidc_groups_prefix", cluster.OIDCGroupsPrefix)
	set("oidc_required_claim", cluster.OIDCRequiredClaim)
	set("oidc_username_prefix", cluster.OIDCUsernamePrefix)
	set("maintenance_window_start_time", cluster.MaintenanceWindowStart)
	set("maintenance_window_duration", cluster.MaintenanceWindowDuration)
	set("state", cluster.State.ID)
	set("state_text", cluster.State.Text)

	if cluster.InternalIPv4Prefix != nil {
		set("internal_ipv4_prefix", cluster.InternalIPv4Prefix.Identifier)
	}
	if cluster.ExternalIPv4Prefix != nil {
		set("external_ipv4_prefix", cluster.ExternalIPv4Prefix.Identifier)
	}
	if cluster.ExternalIPv6Prefix != nil {
		set("external_ipv6_prefix", cluster.ExternalIPv6Prefix.Identifier)
	}

	set("apiserver_allowlist", strings.Fields(cluster.APIServerAllowlist))
	return diags
}
