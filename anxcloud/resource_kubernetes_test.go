package anxcloud

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.anx.io/go-anxcloud/pkg/client"
)

func TestKubernetesResourcesExposeUpdatesAndCurrentFields(t *testing.T) {
	clusterResource := resourceKubernetesCluster()
	require.NotNil(t, clusterResource.UpdateContext)
	for _, field := range []string{
		"cni_plugin",
		"enable_persistent_storage",
		"external_ip_families",
		"enable_oidc_authentication",
		"maintenance_window_start_time",
		"maintenance_window_duration",
		"state",
		"state_text",
		"wait_until_ready",
	} {
		assert.Contains(t, clusterResource.Schema, field)
	}
	assert.False(t, clusterResource.Schema["name"].ForceNew)
	assert.False(t, clusterResource.Schema["enable_autoscaling"].ForceNew)
	assert.NotContains(t, dataSourceKubernetesCluster().Schema, "wait_until_ready")

	nodePoolResource := resourceKubernetesNodePool()
	require.NotNil(t, nodePoolResource.UpdateContext)
	for _, field := range []string{
		"sync_source",
		"cpu_performance_type",
		"autoscaler_enabled",
		"autoscaler_min_nodes",
		"autoscaler_max_nodes",
		"additional_disks",
		"networks",
		"ssh_public_keys",
		"taints",
		"labels",
		"annotations",
		"state",
		"state_text",
	} {
		assert.Contains(t, nodePoolResource.Schema, field)
	}
	assert.False(t, nodePoolResource.Schema["initial_replicas"].ForceNew)
	assert.False(t, nodePoolResource.Schema["disk"].ForceNew)
}

func TestKubernetesClusterCreateDefinitionExcludesReadOnlyFields(t *testing.T) {
	d := schema.TestResourceDataRaw(t, schemaKubernetesCluster(), map[string]any{
		"name":     "test-cluster",
		"location": "location-id",
	})

	definition := kubernetesClusterCreateDefinition(d)

	assert.NotContains(t, definition, "id")
	assert.NotContains(t, definition, "patch_version")
	assert.NotContains(t, definition, "state")
	assert.NotContains(t, definition, "state_text")
	assert.NotContains(t, definition, "wait_until_ready")
}

func TestKubernetesClusterCreateUpdateDefinitionExcludesUnconfiguredFields(t *testing.T) {
	d := schema.TestResourceDataRaw(t, schemaKubernetesCluster(), map[string]any{
		"name":     "test-cluster",
		"location": "location-id",
	})

	definition := kubernetesClusterCreateUpdateDefinitionWithConfiguredFields(d, func(string) bool {
		return false
	})

	for _, field := range []string{
		"cni_plugin",
		"enable_persistent_storage",
		"external_ip_families",
		"enable_oidc_authentication",
		"maintenance_window_start_time",
		"maintenance_window_duration",
	} {
		assert.NotContains(t, definition, field)
	}
}

func TestKubernetesClusterV2FieldsArePatchedAfterCreate(t *testing.T) {
	d := schema.TestResourceDataRaw(t, schemaKubernetesCluster(), map[string]any{
		"name":                      "test-cluster",
		"location":                  "location-id",
		"enable_persistent_storage": true,
		"external_ip_families":      "IPv4",
	})

	createDefinition := kubernetesClusterCreateDefinition(d)
	configuredFields := map[string]bool{
		"enable_persistent_storage": true,
		"external_ip_families":      true,
	}
	updateDefinition := kubernetesClusterCreateUpdateDefinitionWithConfiguredFields(d, func(field string) bool {
		return configuredFields[field]
	})

	assert.NotContains(t, createDefinition, "enable_persistent_storage")
	assert.NotContains(t, createDefinition, "external_ip_families")
	assert.Equal(t, true, updateDefinition["enable_persistent_storage"])
	assert.Equal(t, "IPv4", updateDefinition["external_ip_families"])
}

func TestKubernetesClusterReadPreservesFailedResource(t *testing.T) {
	cli := kubernetesTestClient{do: func(r *http.Request) (*http.Response, error) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/kubernetes/v2/cluster/cluster-id", r.URL.Path)
		return kubernetesTestResponse(t, http.StatusOK, map[string]any{
			"identifier": "cluster-id",
			"name":       "failed-cluster",
			"state": map[string]any{
				"id":    "1",
				"title": "Error",
				"type":  0,
			},
			"version":  map[string]any{"id": "1.34", "title": "1.34"},
			"location": map[string]any{"identifier": "location-id", "name": "ANX04"},
		}), nil
	}}

	d := schema.TestResourceDataRaw(t, schemaKubernetesCluster(), map[string]any{
		"name":     "failed-cluster",
		"location": "location-id",
	})
	d.SetId("cluster-id")

	diags := resourceKubernetesClusterRead(context.Background(), d, providerContext{legacyClient: cli})
	require.False(t, diags.HasError(), "refreshing an error-state cluster must remain usable")
	assert.Equal(t, "cluster-id", d.Id())
	assert.Equal(t, "1", d.Get("state"))
	assert.Equal(t, "Error", d.Get("state_text"))
}

func TestKubernetesClusterCreateDoesNotWaitByDefault(t *testing.T) {
	requestCount := 0
	cli := kubernetesTestClient{do: func(r *http.Request) (*http.Response, error) {
		requestCount++
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/api/kubernetes/v2/cluster", r.URL.Path)
		return kubernetesTestResponse(t, http.StatusOK, map[string]any{
			"identifier": "cluster-id",
			"name":       "test-cluster",
			"location":   map[string]any{"identifier": "location-id", "name": "ANX04"},
			"version":    map[string]any{"id": "1.35", "title": "1.35"},
			"state":      map[string]any{"id": "1", "title": "Error", "type": 1},
		}), nil
	}}
	d := schema.TestResourceDataRaw(t, resourceKubernetesCluster().Schema, map[string]any{
		"name":     "test-cluster",
		"location": "location-id",
	})

	diags := resourceKubernetesClusterCreate(context.Background(), d, providerContext{legacyClient: cli})

	require.False(t, diags.HasError())
	assert.Equal(t, "cluster-id", d.Id())
	assert.Equal(t, "1", d.Get("state"))
	assert.Equal(t, "Error", d.Get("state_text"))
	assert.Equal(t, 1, requestCount)
}

func TestKubernetesNodePoolCreateUpdateDefinitionExcludesUnconfiguredFields(t *testing.T) {
	d := schema.TestResourceDataRaw(t, schemaKubernetesNodePool(), map[string]any{
		"name":             "test-node-pool",
		"initial_replicas": 3,
		"cpus":             2,
		"memory_gib":       4,
		"disk": []any{map[string]any{
			"size_gib": 20,
		}},
		"operating_system": "Flatcar Linux",
		"cluster":          "cluster-id",
	})

	definition := kubernetesNodePoolCreateUpdateDefinitionWithConfiguredFields(d, func(string) bool {
		return false
	})

	for _, field := range []string{
		"syncsource",
		"cpu_performance_type",
		"autoscaler_enabled",
		"autoscaler_min_nodes",
		"autoscaler_max_nodes",
		"dns_override_ipv4",
		"dns_v4_1",
		"dns_v4_2",
		"dns_override_ipv6",
		"dns_v6_1",
		"dns_v6_2",
		"sshpubkeys",
		"taints",
		"labels",
		"annotations",
	} {
		assert.NotContains(t, definition, field)
	}
}

func TestKubernetesNodePoolV2FieldsArePatchedAfterCreate(t *testing.T) {
	d := schema.TestResourceDataRaw(t, schemaKubernetesNodePool(), map[string]any{
		"name":             "test-node-pool",
		"initial_replicas": 3,
		"cpus":             2,
		"memory_gib":       4,
		"disk": []any{map[string]any{
			"size_gib": 20,
		}},
		"operating_system":   "Flatcar Linux",
		"cluster":            "cluster-id",
		"autoscaler_enabled": true,
		"dns_ipv4_1":         "192.0.2.53",
	})

	createDefinition := kubernetesNodePoolCreateDefinition(d)
	configuredFields := map[string]bool{
		"autoscaler_enabled": true,
		"dns_ipv4_1":         true,
	}
	updateDefinition := kubernetesNodePoolCreateUpdateDefinitionWithConfiguredFields(d, func(field string) bool {
		return configuredFields[field]
	})

	assert.NotContains(t, createDefinition, "autoscaler_enabled")
	assert.NotContains(t, createDefinition, "dns_v4_1")
	assert.Equal(t, true, updateDefinition["autoscaler_enabled"])
	assert.Equal(t, "192.0.2.53", updateDefinition["dns_v4_1"])
	assert.NotContains(t, updateDefinition, "syncsource")
}

func TestAwaitKubernetesClusterReconciliationIgnoresIntermediateStates(t *testing.T) {
	states := []map[string]any{
		{"id": "1", "title": "Error", "type": 1},
		{"id": "2", "title": "Waiting for control plane", "type": 0},
		{"id": "0", "title": "Deployed", "type": 1},
	}
	requestCount := 0
	cli := kubernetesTestClient{do: func(r *http.Request) (*http.Response, error) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/api/kubernetes/v2/cluster/cluster-id", r.URL.Path)
		require.Less(t, requestCount, len(states))

		state := states[requestCount]
		requestCount++
		return kubernetesTestResponse(t, http.StatusOK, map[string]any{
			"identifier": "cluster-id",
			"state":      state,
		}), nil
	}}

	a := newKubernetesServiceAPI[kubernetesCluster](cli, "cluster")
	cluster, err := awaitKubernetesClusterReconciliationWithPollInterval(
		context.Background(),
		a,
		"cluster-id",
		0,
	)

	require.NoError(t, err)
	assert.Equal(t, "0", cluster.State.ID)
	assert.Equal(t, len(states), requestCount)
}

func TestKubernetesClusterDeleteSkipsQueuedRulesAndWaitsForNotFound(t *testing.T) {
	cli := kubernetesTestClient{do: func(r *http.Request) (*http.Response, error) {
		switch r.Method {
		case http.MethodDelete:
			assert.Equal(t, "/api/kubernetes/v2/cluster/cluster-id", r.URL.Path)
			assert.Equal(t, "true", r.URL.Query().Get("skip_queued_automation_rules"))
			return kubernetesTestResponse(t, http.StatusNoContent, nil), nil
		case http.MethodGet:
			response := kubernetesTestResponse(t, http.StatusNotFound, map[string]any{
				"error": map[string]any{
					"code":    http.StatusNotFound,
					"message": "not found",
				},
			})
			responseError := &client.ResponseError{Request: r, Response: response}
			responseError.ErrorData.Code = http.StatusNotFound
			responseError.ErrorData.Message = "not found"
			return response, responseError
		default:
			t.Fatalf("unexpected method %s", r.Method)
			return nil, nil
		}
	}}

	d := schema.TestResourceDataRaw(t, schemaKubernetesCluster(), map[string]any{
		"name":     "failed-cluster",
		"location": "location-id",
	})
	d.SetId("cluster-id")

	diags := resourceKubernetesClusterDelete(context.Background(), d, providerContext{legacyClient: cli})
	require.False(t, diags.HasError())
	assert.Empty(t, d.Id())
}

func TestKubernetesPatchPreservesExplicitZeroValues(t *testing.T) {
	cli := kubernetesTestClient{do: func(r *http.Request) (*http.Response, error) {
		assert.Equal(t, http.MethodPatch, r.Method)
		assert.Equal(t, "/api/kubernetes/v2/node_pool/node-pool-id", r.URL.Path)

		var body map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Contains(t, body, "autoscaler_enabled")
		assert.Equal(t, false, body["autoscaler_enabled"])
		assert.Contains(t, body, "replicas")
		assert.Equal(t, float64(0), body["replicas"])

		return kubernetesTestResponse(t, http.StatusOK, map[string]any{
			"identifier": "node-pool-id",
			"state": map[string]any{
				"id":    "0",
				"title": "Deployed",
				"type":  1,
			},
		}), nil
	}}

	a := newKubernetesServiceAPI[kubernetesNodePool](cli, "node_pool")
	_, err := a.Update(context.Background(), "node-pool-id", map[string]any{
		"autoscaler_enabled": false,
		"replicas":           0,
	})
	require.NoError(t, err)
}

type kubernetesTestClient struct {
	do func(*http.Request) (*http.Response, error)
}

func (c kubernetesTestClient) BaseURL() string {
	return "https://engine.test.invalid"
}

func (c kubernetesTestClient) Do(request *http.Request) (*http.Response, error) {
	return c.do(request)
}

func kubernetesTestResponse(t *testing.T, status int, value any) *http.Response {
	t.Helper()
	var body bytes.Buffer
	if value != nil {
		require.NoError(t, json.NewEncoder(&body).Encode(value))
	}
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewReader(body.Bytes())),
	}
}
