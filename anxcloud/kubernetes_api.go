package anxcloud

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"

	"go.anx.io/go-anxcloud/pkg/apis/common/gs"
	"go.anx.io/go-anxcloud/pkg/client"
)

const kubernetesAPIV2Path = "api/kubernetes/v2"

type kubernetesServiceAPI[T any] struct {
	client       client.Client
	resourceName string
}

func newKubernetesServiceAPI[T any](c client.Client, resourceName string) kubernetesServiceAPI[T] {
	return kubernetesServiceAPI[T]{client: c, resourceName: resourceName}
}

func (a kubernetesServiceAPI[T]) Create(ctx context.Context, definition map[string]any) (T, error) {
	var result T
	if err := a.do(ctx, http.MethodPost, "", nil, definition, &result); err != nil {
		return result, fmt.Errorf("failed creating Kubernetes %s: %w", a.resourceName, err)
	}
	return result, nil
}

func (a kubernetesServiceAPI[T]) Get(ctx context.Context, identifier string) (T, error) {
	var result T
	if err := a.do(ctx, http.MethodGet, identifier, nil, nil, &result); err != nil {
		return result, fmt.Errorf("failed getting Kubernetes %s %q: %w", a.resourceName, identifier, err)
	}
	return result, nil
}

func (a kubernetesServiceAPI[T]) Update(ctx context.Context, identifier string, definition map[string]any) (T, error) {
	var result T
	if err := a.do(ctx, http.MethodPatch, identifier, nil, definition, &result); err != nil {
		return result, fmt.Errorf("failed updating Kubernetes %s %q: %w", a.resourceName, identifier, err)
	}
	return result, nil
}

func (a kubernetesServiceAPI[T]) Delete(ctx context.Context, identifier string) error {
	query := url.Values{}
	query.Set("skip_queued_automation_rules", "true")
	if err := a.do(ctx, http.MethodDelete, identifier, query, nil, nil); err != nil {
		return fmt.Errorf("failed deleting Kubernetes %s %q: %w", a.resourceName, identifier, err)
	}
	return nil
}

func (a kubernetesServiceAPI[T]) do(
	ctx context.Context,
	method string,
	identifier string,
	query url.Values,
	body any,
	result any,
) error {
	endpoint, err := url.Parse(a.client.BaseURL())
	if err != nil {
		return fmt.Errorf("parse API base URL: %w", err)
	}

	endpoint.Path = path.Join(endpoint.Path, kubernetesAPIV2Path, a.resourceName, identifier)
	endpoint.RawQuery = query.Encode()

	var requestBody *bytes.Reader
	if body == nil {
		requestBody = bytes.NewReader(nil)
	} else {
		encoded, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encode request body: %w", err)
		}
		requestBody = bytes.NewReader(encoded)
	}

	request, err := http.NewRequestWithContext(ctx, method, endpoint.String(), requestBody)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	response, err := a.client.Do(request)
	if err != nil {
		if response != nil && response.Body != nil {
			response.Body.Close()
		}
		return err
	}
	defer response.Body.Close()

	if result == nil || response.StatusCode == http.StatusNoContent {
		return nil
	}

	if err := json.NewDecoder(response.Body).Decode(result); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

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

type kubernetesResourceReference struct {
	Identifier string `json:"identifier"`
	Name       string `json:"name"`
}

type kubernetesCluster struct {
	Identifier string   `json:"identifier"`
	Name       string   `json:"name"`
	State      gs.State `json:"state"`

	Version         kubernetesSelect            `json:"version"`
	PatchVersion    string                      `json:"patch_version"`
	Location        kubernetesResourceReference `json:"location"`
	NeedsServiceVMs bool                        `json:"needs_service_vms"`
	EnableNAT       bool                        `json:"enable_nat_gateways"`
	EnableLBaaS     bool                        `json:"enable_lbaas"`

	InternalIPv4Prefix *kubernetesResourceReference `json:"internal_ipv4_prefix"`
	ExternalIPv4Prefix *kubernetesResourceReference `json:"external_ipv4_prefix"`
	ExternalIPv6Prefix *kubernetesResourceReference `json:"external_ipv6_prefix"`

	CNIPlugin               kubernetesSelect `json:"cni_plugin"`
	EnablePersistentStorage bool             `json:"enable_persistent_storage"`
	EnableAutoscaling       bool             `json:"autoscaling"`
	APIServerAllowlist      string           `json:"apiserver_allowlist"`
	ExternalIPFamilies      kubernetesSelect `json:"external_ip_families"`

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
}

type kubernetesNodePoolDisk struct {
	Identifier      string           `json:"identifier"`
	Name            string           `json:"name"`
	SizeBytes       int64            `json:"size_bytes"`
	PerformanceType kubernetesSelect `json:"performance_type"`
}

type kubernetesNodePoolNetwork struct {
	Identifier     string                      `json:"identifier"`
	Name           string                      `json:"name"`
	BandwidthLimit kubernetesSelect            `json:"bandwidth_limit"`
	VLAN           kubernetesResourceReference `json:"vlan"`
}

type kubernetesNodePool struct {
	Identifier string   `json:"identifier"`
	Name       string   `json:"name"`
	State      gs.State `json:"state"`

	Replicas           int                         `json:"replicas"`
	CPUs               int                         `json:"cpus"`
	MemoryBytes        int64                       `json:"memory"`
	DiskSizeBytes      int64                       `json:"disk_size"`
	OperatingSystem    kubernetesSelect            `json:"operating_system"`
	Cluster            kubernetesResourceReference `json:"cluster"`
	SyncSource         kubernetesSelect            `json:"syncsource"`
	CPUPerformanceType kubernetesSelect            `json:"cpu_performance_type"`

	AutoscalerEnabled  bool `json:"autoscaler_enabled"`
	AutoscalerMinNodes int  `json:"autoscaler_min_nodes"`
	AutoscalerMaxNodes int  `json:"autoscaler_max_nodes"`

	DiskPerformanceType kubernetesSelect            `json:"disk_performance_type"`
	AdditionalDisks     []kubernetesNodePoolDisk    `json:"additional_disks"`
	Networks            []kubernetesNodePoolNetwork `json:"networks"`

	DNSOverrideIPv4 bool   `json:"dns_override_ipv4"`
	DNSv4Entry1     string `json:"dns_v4_1"`
	DNSv4Entry2     string `json:"dns_v4_2"`
	DNSOverrideIPv6 bool   `json:"dns_override_ipv6"`
	DNSv6Entry1     string `json:"dns_v6_1"`
	DNSv6Entry2     string `json:"dns_v6_2"`

	SSHPublicKeys string `json:"sshpubkeys"`
	Taints        string `json:"taints"`
	Labels        string `json:"labels"`
	Annotations   string `json:"annotations"`
}
