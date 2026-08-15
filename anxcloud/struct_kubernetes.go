package anxcloud

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"go.anx.io/go-anxcloud/pkg/api/types"
	"go.anx.io/go-anxcloud/pkg/apis/common"
	"go.anx.io/go-anxcloud/pkg/apis/common/gs"
	"go.anx.io/go-anxcloud/pkg/kubernetes/network"
	"go.anx.io/go-anxcloud/pkg/kubernetes/nodepool"
)

const gibiFactor = 1073741824 // math.Pow(2, 30)

const kubernetesAPIV2Path = "/api/kubernetes/v2"

// kubernetesSelect is used by the Kubernetes v2 API for selectable values.
// Some older deployments return only the ID, so both representations are accepted.
type kubernetesSelect struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

func (s *kubernetesSelect) UnmarshalJSON(data []byte) error {
	var id string
	if err := json.Unmarshal(data, &id); err == nil {
		s.ID = id
		s.Title = id
		return nil
	}

	type selectAlias kubernetesSelect
	var value selectAlias
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*s = kubernetesSelect(value)
	return nil
}

// kubernetesCluster implements the go-anxcloud generic API object hooks for
// the Kubernetes v2 cluster endpoint. The resource-specific SDK cluster API
// still targets Kubernetes v1 in go-anxcloud v0.14.5.
type kubernetesCluster struct {
	Identifier string   `json:"identifier" anxcloud:"identifier"`
	Name       string   `json:"name"`
	State      gs.State `json:"state"`

	Version         kubernetesSelect       `json:"version"`
	PatchVersion    string                 `json:"patch_version"`
	Location        common.PartialResource `json:"location"`
	NeedsServiceVMs bool                   `json:"needs_service_vms"`
	EnableNAT       bool                   `json:"enable_nat_gateways"`
	EnableLBaaS     bool                   `json:"enable_lbaas"`

	InternalIPv4Prefix *common.PartialResource `json:"internal_ipv4_prefix"`
	ExternalIPv4Prefix *common.PartialResource `json:"external_ipv4_prefix"`
	ExternalIPv6Prefix *common.PartialResource `json:"external_ipv6_prefix"`

	CNIPlugin          kubernetesSelect `json:"cni_plugin"`
	EnableAutoscaling  bool             `json:"autoscaling"`
	APIServerAllowlist string           `json:"apiserver_allowlist"`
	ExternalIPFamilies kubernetesSelect `json:"external_ip_families"`

	EnableOIDCAuthentication bool   `json:"enable_oidc_authentication"`
	OIDCClientID             string `json:"oidc_client_id"`
	OIDCIssuerURL            string `json:"oidc_issuer_url"`
	OIDCGroupsClaim          string `json:"oidc_groups_claim"`
	OIDCUsernameClaim        string `json:"oidc_username_claim"`
	OIDCExtraScopes          string `json:"oidc_extra_scopes"`
	OIDCGroupsPrefix         string `json:"oidc_groups_prefix"`
	OIDCRequiredClaim        string `json:"oidc_required_claim"`
	OIDCUsernamePrefix       string `json:"oidc_username_prefix"`

	MaintenanceWindowStart    string `json:"maintenance_window_start_time"`
	MaintenanceWindowDuration string `json:"maintenance_window_duration"`

	requestDefinition map[string]any
}

// kubernetesNodePool reuses the go-anxcloud Kubernetes v2 response model. Its
// request body remains a map so PATCH can retain explicit false and zero values.
type kubernetesNodePool struct {
	nodepool.Nodepool
	requestDefinition map[string]any
}

// kubernetesNodePoolNetwork reuses the go-anxcloud network model and adds the
// state and parent node-pool fields exposed by the current Kubernetes v2 API.
type kubernetesNodePoolNetwork struct {
	network.NodepoolNetwork
	State    gs.State               `json:"state"`
	NodePool common.PartialResource `json:"nodepool"`

	requestDefinition map[string]any
}

func (c *kubernetesCluster) EndpointURL(context.Context) (*url.URL, error) {
	return url.Parse(kubernetesAPIV2Path + "/cluster")
}

func (c *kubernetesCluster) GetIdentifier(context.Context) (string, error) {
	return c.Identifier, nil
}

func (c *kubernetesCluster) FilterAPIRequestBody(context.Context) (interface{}, error) {
	return c.requestDefinition, nil
}

func (c *kubernetesCluster) FilterAPIRequest(ctx context.Context, request *http.Request) (*http.Request, error) {
	return filterKubernetesAPIRequest(ctx, request)
}

func (n *kubernetesNodePool) EndpointURL(context.Context) (*url.URL, error) {
	return url.Parse(kubernetesAPIV2Path + "/node_pool")
}

func (n *kubernetesNodePool) GetIdentifier(context.Context) (string, error) {
	return n.Identifier, nil
}

func (n *kubernetesNodePool) FilterAPIRequestBody(context.Context) (interface{}, error) {
	return n.requestDefinition, nil
}

func (n *kubernetesNodePool) FilterAPIRequest(ctx context.Context, request *http.Request) (*http.Request, error) {
	return filterKubernetesAPIRequest(ctx, request)
}

func (n *kubernetesNodePoolNetwork) EndpointURL(context.Context) (*url.URL, error) {
	return url.Parse(kubernetesAPIV2Path + "/node_pool_network")
}

func (n *kubernetesNodePoolNetwork) GetIdentifier(context.Context) (string, error) {
	return n.Identifier, nil
}

func (n *kubernetesNodePoolNetwork) FilterAPIRequestBody(context.Context) (interface{}, error) {
	return n.requestDefinition, nil
}

func (n *kubernetesNodePoolNetwork) FilterAPIRequest(ctx context.Context, request *http.Request) (*http.Request, error) {
	return filterKubernetesAPIRequest(ctx, request)
}

func filterKubernetesAPIRequest(ctx context.Context, request *http.Request) (*http.Request, error) {
	operation, err := types.OperationFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if operation == types.OperationUpdate {
		request.Method = http.MethodPatch
	}
	if operation == types.OperationDestroy {
		query := request.URL.Query()
		query.Set("skip_queued_automation_rules", "true")
		request.URL.RawQuery = query.Encode()
	}

	return request, nil
}

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
