package anxcloud

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.anx.io/go-anxcloud/pkg/api"
	"go.anx.io/go-anxcloud/pkg/client"
)

func TestKubernetesResourcesExposeUpdatesAndCurrentFields(t *testing.T) {
	clusterResource := resourceKubernetesCluster()
	require.NotNil(t, clusterResource.UpdateContext)
	for _, field := range []string{
		"cni_plugin",
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
	for fieldName, fieldSchema := range clusterResource.Schema {
		if fieldName == "name" || fieldName == "api_environment" {
			assert.True(t, fieldSchema.ForceNew, "%s identifies a cluster in a specific Kubernetes service", fieldName)
			continue
		}
		assert.False(t, fieldSchema.ForceNew, "%s must be updated without replacing the cluster", fieldName)
	}
	assert.NotContains(t, dataSourceKubernetesCluster().Schema, "wait_until_ready")
	for _, value := range []string{"IPv4", "Dualstack"} {
		_, errors := clusterResource.Schema["external_ip_families"].ValidateFunc(value, "external_ip_families")
		assert.Empty(t, errors, "%s must be a valid external IP family", value)
	}
	_, errors := clusterResource.Schema["external_ip_families"].ValidateFunc("DualStack", "external_ip_families")
	assert.NotEmpty(t, errors, "external IP family values must match the API spelling")

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
	assert.True(t, nodePoolResource.Schema["networks"].Required)
	assert.Equal(t, 1, nodePoolResource.Schema["networks"].MinItems)
	assert.Equal(t, 10, nodePoolResource.Schema["networks"].MaxItems)
	assert.Equal(t, "engine", nodePoolResource.Schema["sync_source"].Default)
	_, errors = nodePoolResource.Schema["sync_source"].ValidateFunc("Cluster", "sync_source")
	assert.Empty(t, errors)
	assert.Equal(t, "cluster", nodePoolResource.Schema["sync_source"].StateFunc("Cluster"))

	for _, resource := range []*schema.Resource{
		clusterResource,
		nodePoolResource,
		resourceKubernetesKubeconfig(),
	} {
		field := resource.Schema["api_environment"]
		require.NotNil(t, field)
		assert.Equal(t, kubernetesAPIEnvironmentProd, field.Default)
		assert.True(t, field.ForceNew)
		for _, environment := range []string{
			kubernetesAPIEnvironmentProd,
			kubernetesAPIEnvironmentStage,
			kubernetesAPIEnvironmentDev,
		} {
			_, validationErrors := field.ValidateFunc(environment, "api_environment")
			assert.Empty(t, validationErrors)
		}
		_, validationErrors := field.ValidateFunc("testing", "api_environment")
		assert.NotEmpty(t, validationErrors)
	}
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
	assert.NotContains(t, definition, "backend_name")
	assert.NotContains(t, definition, "state_text")
	assert.NotContains(t, definition, "wait_until_ready")
	assert.NotContains(t, definition, "api_environment")
}

func TestKubernetesAPIEnvironmentPaths(t *testing.T) {
	tests := []struct {
		environment string
		service     string
	}{
		{environment: kubernetesAPIEnvironmentProd, service: "kubernetes"},
		{environment: kubernetesAPIEnvironmentStage, service: "kubernetes-stage"},
		{environment: kubernetesAPIEnvironmentDev, service: "kubernetes-dev"},
	}

	for _, test := range tests {
		t.Run(test.environment, func(t *testing.T) {
			clusterURL, err := (&kubernetesCluster{apiEnvironment: test.environment}).EndpointURL(context.Background())
			require.NoError(t, err)
			assert.Equal(t, "/api/"+test.service+"/v2/cluster", clusterURL.Path)

			nodePoolURL, err := (&kubernetesNodePool{apiEnvironment: test.environment}).EndpointURL(context.Background())
			require.NoError(t, err)
			assert.Equal(t, "/api/"+test.service+"/v2/node_pool", nodePoolURL.Path)

			kubeconfigClusterURL, err := (&kubernetesKubeconfigCluster{apiEnvironment: test.environment}).EndpointURL(context.Background())
			require.NoError(t, err)
			assert.Equal(t, "/api/"+test.service+"/v1/cluster.json", kubeconfigClusterURL.Path)
		})
	}
}

func TestKubernetesKubeconfigEnvironmentRequestContract(t *testing.T) {
	requestCount := 0
	provider := kubernetesTestProviderContext(t, func(r *http.Request) (*http.Response, error) {
		requestCount++
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Zero(t, r.ContentLength)
		switch requestCount {
		case 1:
			assert.Equal(t,
				kubernetesAPIPath(kubernetesAPIEnvironmentStage, 2)+"/cluster.json/cluster-id/trigger/rotate_kubeconfig",
				r.URL.Path,
			)
		case 2:
			assert.Equal(t,
				kubernetesAPIPath(kubernetesAPIEnvironmentStage, 2)+"/cluster.json/cluster-id/trigger/rotate_kubeconfig",
				r.URL.Path,
			)
		default:
			t.Fatalf("unexpected request %d", requestCount)
		}
		return kubernetesTestResponse(t, http.StatusOK, map[string]any{}), nil
	})

	request := kubernetesKubeconfigRequest{
		Cluster:        "cluster-id",
		apiEnvironment: kubernetesAPIEnvironmentStage,
	}
	require.NoError(t, provider.api.Create(context.Background(), &request))
	require.NoError(t, provider.api.Destroy(context.Background(), &request))
	assert.Equal(t, 2, requestCount)
}

func TestGetKubernetesKubeconfigUsesSelectedEnvironment(t *testing.T) {
	provider := kubernetesTestProviderContext(t, func(r *http.Request) (*http.Response, error) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t,
			kubernetesAPIPath(kubernetesAPIEnvironmentDev, 1)+"/cluster.json/cluster-id",
			r.URL.Path,
		)
		return kubernetesTestResponse(t, http.StatusOK, map[string]any{
			"identifier": "cluster-id",
			"kubeconfig": "test-kubeconfig",
		}), nil
	})

	raw, err := getKubernetesKubeconfig(
		context.Background(),
		provider.api,
		"cluster-id",
		kubernetesAPIEnvironmentDev,
	)
	require.NoError(t, err)
	assert.Equal(t, "test-kubeconfig", raw)
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
		"enable_oidc_authentication",
		"maintenance_window_start_time",
		"maintenance_window_duration",
	} {
		assert.NotContains(t, definition, field)
	}
}

func TestKubernetesClusterOIDCClaimsAreNullable(t *testing.T) {
	d := schema.TestResourceDataRaw(t, schemaKubernetesCluster(), map[string]any{
		"name":                       "test-cluster",
		"location":                   "location-id",
		"enable_oidc_authentication": false,
	})

	configuredFields := map[string]bool{
		"enable_oidc_authentication": true,
	}
	definition := kubernetesClusterCreateUpdateDefinitionWithConfiguredFields(d, func(field string) bool {
		return configuredFields[field]
	})

	assert.Contains(t, definition, "oidc_groups_claim")
	assert.Nil(t, definition["oidc_groups_claim"])
	assert.Contains(t, definition, "oidc_username_claim")
	assert.Nil(t, definition["oidc_username_claim"])
	encodedDefinition, err := json.Marshal(definition)
	require.NoError(t, err)
	assert.JSONEq(t, `{
		"enable_oidc_authentication": false,
		"oidc_groups_claim": null,
		"oidc_username_claim": null
	}`, string(encodedDefinition))

	d = schema.TestResourceDataRaw(t, schemaKubernetesCluster(), map[string]any{
		"name":                "test-cluster",
		"location":            "location-id",
		"oidc_groups_claim":   "groups",
		"oidc_username_claim": "preferred_username",
	})
	configuredFields = map[string]bool{
		"oidc_groups_claim":   true,
		"oidc_username_claim": true,
	}
	definition = kubernetesClusterCreateUpdateDefinitionWithConfiguredFields(d, func(field string) bool {
		return configuredFields[field]
	})

	assert.Equal(t, "groups", definition["oidc_groups_claim"])
	assert.Equal(t, "preferred_username", definition["oidc_username_claim"])

	d = schema.TestResourceDataRaw(t, schemaKubernetesCluster(), map[string]any{})
	assert.Nil(t, kubernetesNullableOIDCClaim(d, "oidc_groups_claim"))
	assert.Nil(t, kubernetesNullableOIDCClaim(d, "oidc_username_claim"))
}

func TestKubernetesClusterFieldsUseCorrectCreateAndPatchRequests(t *testing.T) {
	d := schema.TestResourceDataRaw(t, schemaKubernetesCluster(), map[string]any{
		"name":                 "test-cluster",
		"location":             "location-id",
		"cni_plugin":           "canal",
		"external_ip_families": "IPv4",
	})

	createDefinition := kubernetesClusterCreateDefinition(d)
	configuredFields := map[string]bool{
		"cni_plugin":           true,
		"external_ip_families": true,
	}
	updateDefinition := kubernetesClusterCreateUpdateDefinitionWithConfiguredFields(d, func(field string) bool {
		return configuredFields[field]
	})

	assert.NotContains(t, createDefinition, "cni_plugin")
	assert.Equal(t, "IPv4", createDefinition["external_ip_families"])
	assert.Equal(t, "canal", updateDefinition["cni_plugin"])
	assert.NotContains(t, updateDefinition, "external_ip_families")
}

func TestKubernetesClusterReadPreservesFailedResource(t *testing.T) {
	provider := kubernetesTestProviderContext(t, func(r *http.Request) (*http.Response, error) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, kubernetesAPIPath(kubernetesAPIEnvironmentProd, 2)+"/cluster/cluster-id", r.URL.Path)
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
	})

	d := schema.TestResourceDataRaw(t, schemaKubernetesCluster(), map[string]any{
		"name":     "failed-cluster",
		"location": "location-id",
	})
	d.SetId("cluster-id")

	diags := resourceKubernetesClusterRead(context.Background(), d, provider)
	require.False(t, diags.HasError(), "refreshing an error-state cluster must remain usable")
	assert.Equal(t, "cluster-id", d.Id())
	assert.Equal(t, "1", d.Get("state"))
	assert.Equal(t, "Error", d.Get("state_text"))
}

func TestKubernetesClusterCreateDoesNotWaitByDefault(t *testing.T) {
	requestCount := 0
	provider := kubernetesTestProviderContext(t, func(r *http.Request) (*http.Response, error) {
		requestCount++
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, kubernetesAPIPath(kubernetesAPIEnvironmentProd, 2)+"/cluster", r.URL.Path)
		var body map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.NotContains(t, body, "state")
		assert.NotContains(t, body, "backend_name")
		assert.Equal(t, "Dualstack", body["external_ip_families"])
		return kubernetesTestResponse(t, http.StatusOK, map[string]any{
			"identifier": "cluster-id",
			"name":       "test-cluster",
			"location":   map[string]any{"identifier": "location-id", "name": "ANX04"},
			"version":    map[string]any{"id": "1.35", "title": "1.35"},
			"external_ip_families": map[string]any{
				"id":    "Dualstack",
				"title": "IPv4 and IPv6",
			},
			"state": map[string]any{"id": "1", "title": "Error", "type": 1},
		}), nil
	})
	d := schema.TestResourceDataRaw(t, resourceKubernetesCluster().Schema, map[string]any{
		"name":                 "test-cluster",
		"location":             "location-id",
		"external_ip_families": "Dualstack",
	})

	diags := resourceKubernetesClusterCreate(context.Background(), d, provider)

	require.False(t, diags.HasError())
	assert.Equal(t, "cluster-id", d.Id())
	assert.Equal(t, "1", d.Get("state"))
	assert.Equal(t, "Error", d.Get("state_text"))
	assert.Equal(t, "Dualstack", d.Get("external_ip_families"))
	assert.Equal(t, 1, requestCount)
}

func TestKubernetesNodePoolCreateDefinitionIncludesCompleteConfiguration(t *testing.T) {
	d := schema.TestResourceDataRaw(t, schemaKubernetesNodePool(), map[string]any{
		"name":             "test-node-pool",
		"initial_replicas": 3,
		"cpus":             2,
		"memory_gib":       4,
		"disk": []any{map[string]any{
			"size_gib": 20,
		}},
		"operating_system":     "Flatcar Linux",
		"cluster":              "cluster-id",
		"sync_source":          "Cluster",
		"cpu_performance_type": "standard",
		"autoscaler_enabled":   true,
		"autoscaler_min_nodes": 2,
		"autoscaler_max_nodes": 5,
		"dns_override_ipv4":    true,
		"dns_ipv4_1":           "192.0.2.53",
		"dns_override_ipv6":    true,
		"dns_ipv6_1":           "2001:db8::53",
		"ssh_public_keys":      "ssh-ed25519 test",
		"taints":               "dedicated=test:NoSchedule",
		"labels":               "role=test",
		"annotations":          "example.com/test=true",
		"networks": []any{map[string]any{
			"name":            "internal",
			"bandwidth_limit": "1000",
			"vlan":            "vlan-id",
		}},
	})

	definition := kubernetesNodePoolCreateDefinition(d)
	assert.Equal(t, "cluster", definition["syncsource"])
	assert.Equal(t, "standard", definition["cpu_performance_type"])
	assert.Equal(t, true, definition["autoscaler_enabled"])
	assert.Equal(t, 2, definition["autoscaler_min_nodes"])
	assert.Equal(t, 5, definition["autoscaler_max_nodes"])
	assert.Equal(t, true, definition["dns_override_ipv4"])
	assert.Equal(t, "192.0.2.53", definition["dns_v4_1"])
	assert.Equal(t, true, definition["dns_override_ipv6"])
	assert.Equal(t, "2001:db8::53", definition["dns_v6_1"])
	assert.Equal(t, "ssh-ed25519 test", definition["sshpubkeys"])
	assert.Equal(t, "dedicated=test:NoSchedule", definition["taints"])
	assert.Equal(t, "role=test", definition["labels"])
	assert.Equal(t, "example.com/test=true", definition["annotations"])
	require.Len(t, definition["networks"], 1)
	assert.Equal(t, "vlan-id", definition["networks"].([]map[string]any)[0]["vlan"])
	assert.NotContains(t, definition, "state")
	assert.NotContains(t, definition, "api_environment")
}

func TestAwaitKubernetesClusterReconciliationIgnoresIntermediateStates(t *testing.T) {
	states := []map[string]any{
		{"id": "1", "title": "Error", "type": 1},
		{"id": "2", "title": "Waiting for control plane", "type": 0},
		{"id": "0", "title": "Deployed", "type": 1},
	}
	requestCount := 0
	provider := kubernetesTestProviderContext(t, func(r *http.Request) (*http.Response, error) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, kubernetesAPIPath(kubernetesAPIEnvironmentProd, 2)+"/cluster/cluster-id", r.URL.Path)
		require.Less(t, requestCount, len(states))

		state := states[requestCount]
		requestCount++
		return kubernetesTestResponse(t, http.StatusOK, map[string]any{
			"identifier": "cluster-id",
			"state":      state,
		}), nil
	})

	cluster, err := awaitKubernetesClusterReconciliationWithPollInterval(
		context.Background(),
		provider.api,
		"cluster-id",
		kubernetesAPIEnvironmentProd,
		0,
	)

	require.NoError(t, err)
	assert.Equal(t, "0", cluster.State.ID)
	assert.Equal(t, len(states), requestCount)
}

func TestKubernetesClusterDeleteSkipsQueuedRulesAndWaitsForNotFound(t *testing.T) {
	provider := kubernetesTestProviderContext(t, func(r *http.Request) (*http.Response, error) {
		switch r.Method {
		case http.MethodDelete:
			assert.Equal(t, kubernetesAPIPath(kubernetesAPIEnvironmentProd, 2)+"/cluster/cluster-id", r.URL.Path)
			assert.Equal(t, "true", r.URL.Query().Get("skip_queued_automation_rules"))
			return kubernetesTestResponse(t, http.StatusNoContent, nil), nil
		case http.MethodGet:
			return kubernetesTestResponse(t, http.StatusNotFound, map[string]any{
				"error": map[string]any{
					"code":    http.StatusNotFound,
					"message": "not found",
				},
			}), nil
		default:
			t.Fatalf("unexpected method %s", r.Method)
			return nil, nil
		}
	})

	d := schema.TestResourceDataRaw(t, schemaKubernetesCluster(), map[string]any{
		"name":     "failed-cluster",
		"location": "location-id",
	})
	d.SetId("cluster-id")

	diags := resourceKubernetesClusterDelete(context.Background(), d, provider)
	require.False(t, diags.HasError())
	assert.Empty(t, d.Id())
}

func TestKubernetesPatchPreservesExplicitZeroValues(t *testing.T) {
	provider := kubernetesTestProviderContext(t, func(r *http.Request) (*http.Response, error) {
		assert.Equal(t, http.MethodPatch, r.Method)
		assert.Equal(t, kubernetesAPIPath(kubernetesAPIEnvironmentProd, 2)+"/node_pool/node-pool-id", r.URL.Path)

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
	})

	nodePool := kubernetesNodePool{requestDefinition: map[string]any{
		"autoscaler_enabled": false,
		"replicas":           0,
	}}
	nodePool.Identifier = "node-pool-id"
	err := provider.api.Update(context.Background(), &nodePool)
	require.NoError(t, err)
}

func TestKubernetesNodePoolSDKRequestContract(t *testing.T) {
	requestCount := 0
	provider := kubernetesTestProviderContext(t, func(r *http.Request) (*http.Response, error) {
		requestCount++
		switch requestCount {
		case 1:
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, kubernetesAPIPath(kubernetesAPIEnvironmentProd, 2)+"/node_pool", r.URL.Path)
			var body map[string]any
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			assert.Equal(t, "test-node-pool", body["name"])
			assert.Equal(t, "engine", body["syncsource"])
			assert.Equal(t, "performance", body["cpu_performance_type"])
			assert.Equal(t, false, body["autoscaler_enabled"])
			assert.Equal(t, float64(0), body["autoscaler_min_nodes"])
			assert.Equal(t, float64(0), body["autoscaler_max_nodes"])
			assert.NotContains(t, body, "state")
			return kubernetesTestResponse(t, http.StatusOK, map[string]any{
				"identifier": "node-pool-id",
				"name":       "test-node-pool",
			}), nil
		case 2:
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, kubernetesAPIPath(kubernetesAPIEnvironmentProd, 2)+"/node_pool/node-pool-id", r.URL.Path)
			return kubernetesTestResponse(t, http.StatusOK, map[string]any{
				"identifier": "node-pool-id",
				"name":       "test-node-pool",
			}), nil
		case 3:
			assert.Equal(t, http.MethodDelete, r.Method)
			assert.Equal(t, kubernetesAPIPath(kubernetesAPIEnvironmentProd, 2)+"/node_pool/node-pool-id", r.URL.Path)
			assert.Equal(t, "true", r.URL.Query().Get("skip_queued_automation_rules"))
			return kubernetesTestResponse(t, http.StatusNoContent, nil), nil
		default:
			t.Fatalf("unexpected request %d: %s %s", requestCount, r.Method, r.URL)
			return nil, nil
		}
	})

	nodePool := kubernetesNodePool{requestDefinition: map[string]any{
		"name":                 "test-node-pool",
		"syncsource":           "engine",
		"cpu_performance_type": "performance",
		"autoscaler_enabled":   false,
		"autoscaler_min_nodes": 0,
		"autoscaler_max_nodes": 0,
	}}
	require.NoError(t, provider.api.Create(context.Background(), &nodePool))
	assert.Equal(t, "node-pool-id", nodePool.Identifier)

	loaded, err := getKubernetesNodePool(context.Background(), provider.api, nodePool.Identifier, kubernetesAPIEnvironmentProd)
	require.NoError(t, err)
	assert.Equal(t, "test-node-pool", loaded.Name)

	require.NoError(t, provider.api.Destroy(context.Background(), &loaded))
	assert.Equal(t, 3, requestCount)
}

func TestKubernetesClusterUpdateUsesV2Patch(t *testing.T) {
	provider := kubernetesTestProviderContext(t, func(r *http.Request) (*http.Response, error) {
		assert.Equal(t, http.MethodPatch, r.Method)
		assert.Equal(t, kubernetesAPIPath(kubernetesAPIEnvironmentProd, 2)+"/cluster/cluster-id", r.URL.Path)

		var body map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, "canal", body["cni_plugin"])
		assert.Contains(t, body, "maintenance_window_duration")
		assert.Equal(t, "", body["maintenance_window_duration"])

		return kubernetesTestResponse(t, http.StatusOK, map[string]any{
			"identifier": "cluster-id",
			"state": map[string]any{
				"id":    "2",
				"title": "Reconciling",
				"type":  0,
			},
		}), nil
	})

	cluster := kubernetesCluster{
		Identifier: "cluster-id",
		requestDefinition: map[string]any{
			"cni_plugin":                  "canal",
			"maintenance_window_duration": "",
		},
	}
	require.NoError(t, provider.api.Update(context.Background(), &cluster))
	assert.Equal(t, "2", cluster.State.ID)
}

func TestKubernetesClusterUpdateAcceptsErrorState(t *testing.T) {
	provider := kubernetesTestProviderContext(t, func(r *http.Request) (*http.Response, error) {
		assert.Equal(t, http.MethodPatch, r.Method)
		assert.Equal(t, kubernetesAPIPath(kubernetesAPIEnvironmentProd, 2)+"/cluster/cluster-id", r.URL.Path)

		var body map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.NotContains(t, body, "name")
		assert.Equal(t, "location-id", body["location"])
		assert.NotContains(t, body, "backend_name")

		return kubernetesTestResponse(t, http.StatusOK, map[string]any{
			"identifier": "cluster-id",
			"name":       "test-cluster",
			"location":   map[string]any{"identifier": "location-id", "name": "ANX04"},
			"version":    map[string]any{"id": "1.35", "title": "1.35"},
			"state":      map[string]any{"id": "1", "title": "Error", "type": 1},
		}), nil
	})

	d := schema.TestResourceDataRaw(t, resourceKubernetesCluster().Schema, map[string]any{
		"name":     "test-cluster",
		"location": "location-id",
	})
	d.SetId("cluster-id")

	diags := resourceKubernetesClusterUpdate(context.Background(), d, provider)
	require.False(t, diags.HasError(), "an error reconciliation state must not reject a configuration update")
	assert.Equal(t, "cluster-id", d.Id())
	assert.Equal(t, "1", d.Get("state"))
	assert.Equal(t, "Error", d.Get("state_text"))
}

type kubernetesTestRoundTripper struct {
	do func(*http.Request) (*http.Response, error)
}

func (r kubernetesTestRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	return r.do(request)
}

func kubernetesTestProviderContext(t *testing.T, do func(*http.Request) (*http.Response, error)) providerContext {
	t.Helper()
	a, err := api.NewAPI(api.WithClientOptions(
		client.IgnoreMissingToken(),
		client.BaseURL("https://engine.test.invalid"),
		client.WithClient(&http.Client{Transport: kubernetesTestRoundTripper{do: do}}),
	))
	require.NoError(t, err)
	return providerContext{api: a}
}

func kubernetesTestResponse(t *testing.T, status int, value any) *http.Response {
	t.Helper()
	var body bytes.Buffer
	if value != nil {
		require.NoError(t, json.NewEncoder(&body).Encode(value))
	}
	return &http.Response{
		Status:     fmt.Sprintf("%d %s", status, http.StatusText(status)),
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewReader(body.Bytes())),
	}
}
