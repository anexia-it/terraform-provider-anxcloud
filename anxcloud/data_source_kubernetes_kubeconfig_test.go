package anxcloud

import (
	"context"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const kubernetesTestKubeconfig = `apiVersion: v1
kind: Config
clusters:
- name: target
  cluster:
    server: https://kubernetes.example.invalid
    certificate-authority-data: Y2E=
users:
- name: admin
  user:
    token: test-token
contexts:
- name: target
  context:
    cluster: target
    user: admin
current-context: target
`

func TestKubernetesKubeconfigDataSourceReadsWithoutMutation(t *testing.T) {
	requestCount := 0
	provider := kubernetesTestProviderContext(t, func(r *http.Request) (*http.Response, error) {
		requestCount++
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t,
			kubernetesAPIPath(kubernetesAPIEnvironmentStage, 1)+"/cluster.json/cluster-id",
			r.URL.Path,
		)
		return kubernetesTestResponse(t, http.StatusOK, map[string]any{
			"identifier": "cluster-id",
			"kubeconfig": kubernetesTestKubeconfig,
		}), nil
	})

	d := schema.TestResourceDataRaw(t, dataSourceKubernetesKubeconfig().Schema, map[string]any{
		"cluster":         "cluster-id",
		"api_environment": kubernetesAPIEnvironmentStage,
	})
	diags := dataSourceKubernetesKubeconfigRead(context.Background(), d, provider)

	require.False(t, diags.HasError())
	assert.Equal(t, 1, requestCount, "the data source must only read the cluster")
	assert.Equal(t, "cluster-id", d.Id())
	assert.Equal(t, "https://kubernetes.example.invalid", d.Get("host"))
	assert.Equal(t, "ca", d.Get("cluster_ca_certificate"))
	assert.Equal(t, "test-token", d.Get("token"))
	assert.Equal(t, kubernetesTestKubeconfig, d.Get("raw"))
}

func TestKubernetesKubeconfigDataSourceDoesNotCreateMissingKubeconfig(t *testing.T) {
	requestCount := 0
	provider := kubernetesTestProviderContext(t, func(r *http.Request) (*http.Response, error) {
		requestCount++
		assert.Equal(t, http.MethodGet, r.Method)
		return kubernetesTestResponse(t, http.StatusOK, map[string]any{
			"identifier": "cluster-id",
		}), nil
	})

	d := schema.TestResourceDataRaw(t, dataSourceKubernetesKubeconfig().Schema, map[string]any{
		"cluster": "cluster-id",
	})
	diags := dataSourceKubernetesKubeconfigRead(context.Background(), d, provider)

	require.True(t, diags.HasError())
	assert.Contains(t, diags[0].Summary, "does not have an existing kubeconfig")
	assert.Equal(t, 1, requestCount, "a missing kubeconfig must not trigger an automation rule")
}

func TestKubernetesKubeconfigDataSourceSchema(t *testing.T) {
	dataSource := dataSourceKubernetesKubeconfig()
	assert.False(t, dataSource.Schema["cluster"].ForceNew)
	assert.False(t, dataSource.Schema["api_environment"].ForceNew)
	assert.Equal(t, kubernetesAPIEnvironmentProd, dataSource.Schema["api_environment"].Default)
	for _, field := range []string{"host", "token", "cluster_ca_certificate", "raw"} {
		assert.True(t, dataSource.Schema[field].Computed)
	}
}
