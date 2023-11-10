package anxcloud

import (
	"fmt"
	"net/http"

	"github.com/onsi/gomega/ghttp"
)

type ghttpMock struct {
	server *ghttp.Server
}

func (m *ghttpMock) appendGetTagsHandler(id string) {
	m.server.AppendHandlers(ghttp.CombineHandlers(
		ghttp.VerifyRequest("GET", fmt.Sprintf("/api/core/v1/resource.json/%s", id)),
		ghttp.RespondWithJSONEncoded(http.StatusOK, map[string]any{
			"tags": []string{},
		}),
	))
}

func (m *ghttpMock) appendCreateClusterHandler() {
	m.server.AppendHandlers(ghttp.CombineHandlers(
		ghttp.VerifyRequest("POST", "/api/kubernetes/v1/cluster.json"),
		ghttp.VerifyJSON(`{"name":"foo","needs_service_vms":true,"enable_nat_gateways":true,"enable_lbaas":true,"location":"test-location"}`),
		ghttp.RespondWithJSONEncoded(http.StatusOK, map[string]any{
			"identifier": "test-cluster-identifier",
		}),
	))
}

func (m *ghttpMock) appendGetClusterHandler() {
	m.server.AppendHandlers(ghttp.CombineHandlers(
		ghttp.VerifyRequest("GET", "/api/kubernetes/v1/cluster.json/test-cluster-identifier"),
		ghttp.RespondWithJSONEncoded(http.StatusOK, map[string]any{
			"identifier":          "test-cluster-identifier",
			"name":                "foo",
			"state":               map[string]any{"type": 1},
			"needs_service_vms":   true,
			"enable_nat_gateways": true,
			"enable_lbaas":        true,
			"location":            map[string]any{"identifier": "test-location"},
		}),
	))
}

func (m *ghttpMock) appendDeleteClusterHandler() {
	m.server.AppendHandlers(ghttp.CombineHandlers(
		ghttp.VerifyRequest("DELETE", "/api/kubernetes/v1/cluster.json/test-cluster-identifier"),
	))
}

func (m *ghttpMock) appendCreateNodePoolHandler() {
	m.server.AppendHandlers(ghttp.CombineHandlers(
		ghttp.VerifyRequest("POST", "/api/kubernetes/v1/node_pool.json"),
		ghttp.VerifyJSON(`{"name":"foo","replicas":3,"cpus":2,"memory":4294967296,"disk_size":21474836480,"operating_system":"Flatcar Linux","cluster":"test-cluster"}`),
		ghttp.RespondWithJSONEncoded(http.StatusOK, map[string]any{
			"identifier": "test-node-pool-identifier",
		}),
	))
}

func (m *ghttpMock) appendGetNodePoolHandler() {
	m.server.AppendHandlers(ghttp.CombineHandlers(
		ghttp.VerifyRequest("GET", "/api/kubernetes/v1/node_pool.json/test-node-pool-identifier"),
		ghttp.RespondWithJSONEncoded(http.StatusOK, map[string]any{
			"identifier":       "test-node-pool-identifier",
			"state":            map[string]any{"type": 1},
			"name":             "foo",
			"replicas":         3,
			"cpus":             2,
			"memory":           4294967296,
			"disk_size":        21474836480,
			"operating_system": "Flatcar Linux",
			"cluster":          map[string]any{"identifier": "test-cluster"},
		}),
	))
}

func (m *ghttpMock) appendDeleteNodePoolHandler() {
	m.server.AppendHandlers(ghttp.CombineHandlers(
		ghttp.VerifyRequest("DELETE", "/api/kubernetes/v1/node_pool.json/test-node-pool-identifier"),
	))
}
