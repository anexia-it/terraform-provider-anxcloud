package anxcloud

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKubernetesNodePoolNetworkResourceSchema(t *testing.T) {
	resource := resourceKubernetesNodePoolNetwork()
	require.NotNil(t, resource.CreateContext)
	require.NotNil(t, resource.ReadContext)
	require.NotNil(t, resource.UpdateContext)
	require.NotNil(t, resource.DeleteContext)

	for _, field := range []string{"name", "node_pool", "bandwidth_limit", "vlan", "state", "state_text"} {
		assert.Contains(t, resource.Schema, field)
		assert.False(t, resource.Schema[field].ForceNew, "%s must be patchable", field)
	}
	for _, field := range []string{"name", "node_pool", "bandwidth_limit", "vlan"} {
		assert.True(t, resource.Schema[field].Required, "%s must be configured for creation", field)
	}

	assert.Contains(t, Provider("test").ResourcesMap, "anxcloud_kubernetes_node_pool_network")
}

func TestKubernetesNodePoolNetworkSDKRequestContract(t *testing.T) {
	requestCount := 0
	provider := kubernetesTestProviderContext(t, func(r *http.Request) (*http.Response, error) {
		requestCount++
		switch requestCount {
		case 1:
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/api/kubernetes/v2/node_pool_network", r.URL.Path)
			assertKubernetesNodePoolNetworkRequestBody(t, r, map[string]any{
				"name":            "storage",
				"nodepool":        "node-pool-id",
				"bandwidth_limit": "1000",
				"vlan":            "vlan-id",
			})
			return kubernetesNodePoolNetworkTestResponse(t, "1000"), nil
		case 2:
			assert.Equal(t, http.MethodPatch, r.Method)
			assert.Equal(t, "/api/kubernetes/v2/node_pool_network/network-id", r.URL.Path)
			assertKubernetesNodePoolNetworkRequestBody(t, r, map[string]any{
				"bandwidth_limit": "10000",
			})
			return kubernetesNodePoolNetworkTestResponse(t, "10000"), nil
		case 3:
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/api/kubernetes/v2/node_pool_network/network-id", r.URL.Path)
			return kubernetesNodePoolNetworkTestResponse(t, "10000"), nil
		case 4:
			assert.Equal(t, http.MethodDelete, r.Method)
			assert.Equal(t, "/api/kubernetes/v2/node_pool_network/network-id", r.URL.Path)
			assert.Equal(t, "true", r.URL.Query().Get("skip_queued_automation_rules"))
			return kubernetesTestResponse(t, http.StatusNoContent, nil), nil
		default:
			t.Fatalf("unexpected request %d: %s %s", requestCount, r.Method, r.URL)
			return nil, nil
		}
	})

	network := kubernetesNodePoolNetwork{requestDefinition: map[string]any{
		"name":            "storage",
		"nodepool":        "node-pool-id",
		"bandwidth_limit": "1000",
		"vlan":            "vlan-id",
	}}
	require.NoError(t, provider.api.Create(context.Background(), &network))
	assert.Equal(t, "network-id", network.Identifier)
	assert.Equal(t, "node-pool-id", network.NodePool.Identifier)
	assert.Equal(t, "1", network.State.ID)

	network.requestDefinition = map[string]any{"bandwidth_limit": "10000"}
	require.NoError(t, provider.api.Update(context.Background(), &network))
	assert.Equal(t, "10000", network.BandwidthLimit.ID)

	loaded, err := getKubernetesNodePoolNetwork(context.Background(), provider.api, network.Identifier)
	require.NoError(t, err)
	assert.Equal(t, "vlan-id", loaded.VLAN.Identifier)
	assert.Equal(t, "Error", loaded.State.Text)

	require.NoError(t, provider.api.Destroy(context.Background(), &loaded))
	assert.Equal(t, 4, requestCount)
}

func TestKubernetesNodePoolReadDoesNotAdoptStandaloneNetworks(t *testing.T) {
	d := schema.TestResourceDataRaw(t, schemaKubernetesNodePool(), map[string]any{
		"name":             "test-node-pool",
		"initial_replicas": 1,
		"cpus":             2,
		"memory_gib":       4,
		"operating_system": "Flatcar Linux",
		"cluster":          "cluster-id",
		"disk": []any{map[string]any{
			"size_gib": 20,
		}},
		"networks": []any{map[string]any{
			"name":            "internal",
			"bandwidth_limit": "1000",
			"vlan":            "internal-vlan-id",
		}},
	})

	var nodePool kubernetesNodePool
	require.NoError(t, json.Unmarshal([]byte(`{
		"networks": [
		{
			"identifier": "internal-network-id",
			"name": "internal",
			"bandwidth_limit": {"id": "1000", "title": "1000 Mbit/s"},
			"vlan": {"identifier": "internal-vlan-id", "name": "VLAN123"}
		},
		{
			"identifier": "external-network-id",
			"name": "external",
			"bandwidth_limit": {"id": "1000", "title": "1000 Mbit/s"},
			"vlan": {"identifier": "external-vlan-id", "name": "VLAN456"}
		}]
	}`), &nodePool))

	require.Empty(t, setResourceDataFromKubernetesNodePool(d, nodePool))
	networks := d.Get("networks").([]any)
	require.Len(t, networks, 1)
	assert.Equal(t, "internal", networks[0].(map[string]any)["name"])
	assert.Equal(t, "internal-vlan-id", networks[0].(map[string]any)["vlan"])
}

func assertKubernetesNodePoolNetworkRequestBody(t *testing.T, r *http.Request, expected map[string]any) {
	t.Helper()
	var body map[string]any
	require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
	assert.Equal(t, expected, body)
}

func kubernetesNodePoolNetworkTestResponse(t *testing.T, bandwidthLimit string) *http.Response {
	t.Helper()
	return kubernetesTestResponse(t, http.StatusOK, map[string]any{
		"identifier": "network-id",
		"name":       "storage",
		"nodepool":   map[string]any{"identifier": "node-pool-id", "name": "test-node-pool"},
		"bandwidth_limit": map[string]any{
			"id":    bandwidthLimit,
			"title": bandwidthLimit + " Mbit/s",
		},
		"vlan":  map[string]any{"identifier": "vlan-id", "name": "VLAN123"},
		"state": map[string]any{"id": "1", "title": "Error", "type": 0},
	})
}
