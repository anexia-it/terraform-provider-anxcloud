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
)

func TestNetworkPrefixDataSourceSchema(t *testing.T) {
	dataSource := dataSourceNetworkPrefix()
	require.NotNil(t, dataSource.ReadContext)

	for _, field := range []string{
		"id",
		"cidr",
		"location_id",
		"netmask",
		"ip_version",
		"type",
		"vlan_id",
		"router_redundancy",
		"description_customer",
		"description_internal",
		"role_text",
		"status",
		"locations",
	} {
		assert.Contains(t, dataSource.Schema, field)
	}

	assert.ElementsMatch(t, []string{"id", "cidr"}, dataSource.Schema["id"].ExactlyOneOf)
	assert.ElementsMatch(t, []string{"id", "cidr"}, dataSource.Schema["cidr"].ExactlyOneOf)
}

func TestNetworkPrefixDataSourceReadByCIDR(t *testing.T) {
	const (
		prefixID = "prefix-id"
		cidr     = "10.0.0.0/24"
	)

	client := networkPrefixDataSourceTestClient{do: func(request *http.Request) (*http.Response, error) {
		require.Equal(t, http.MethodGet, request.Method)

		switch request.URL.Path {
		case "/api/ipam/v1/prefix.json":
			require.Equal(t, cidr, request.URL.Query().Get("search"))
			if request.URL.Query().Get("page") == "1" {
				return networkPrefixDataSourceTestResponse(t, map[string]interface{}{
					"data": []map[string]interface{}{{
						"identifier": prefixID,
						"name":       cidr,
					}},
				}), nil
			}
			return networkPrefixDataSourceTestResponse(t, map[string]interface{}{"data": []interface{}{}}), nil
		case "/api/ipam/v1/prefix.json/" + prefixID:
			return networkPrefixDataSourceTestResponse(t, map[string]interface{}{
				"identifier":           prefixID,
				"name":                 cidr,
				"description_customer": "customer description",
				"description_internal": "internal description",
				"version":              4,
				"netmask":              24,
				"role_text":            "Customer",
				"status":               "Active",
				"router_redundancy":    true,
				"type":                 1,
				"locations": []map[string]interface{}{{
					"identifier": "location-id",
					"code":       "ANX04",
				}},
				"vlans": []map[string]interface{}{{
					"identifier": "vlan-id",
				}},
			}), nil
		default:
			t.Fatalf("unexpected request path %s", request.URL.Path)
			return nil, nil
		}
	}}

	data := schema.TestResourceDataRaw(t, dataSourceNetworkPrefix().Schema, map[string]interface{}{
		"cidr": cidr,
	})
	diags := dataSourceNetworkPrefixRead(context.Background(), data, providerContext{legacyClient: client})

	require.False(t, diags.HasError())
	assert.Equal(t, prefixID, data.Id())
	assert.Equal(t, cidr, data.Get("cidr"))
	assert.Equal(t, "location-id", data.Get("location_id"))
	assert.Equal(t, "vlan-id", data.Get("vlan_id"))
	assert.Equal(t, 4, data.Get("ip_version"))
	assert.Equal(t, 24, data.Get("netmask"))
	assert.Equal(t, 1, data.Get("type"))
	assert.Equal(t, true, data.Get("router_redundancy"))
	assert.Equal(t, "Active", data.Get("status"))
}

type networkPrefixDataSourceTestClient struct {
	do func(*http.Request) (*http.Response, error)
}

func (c networkPrefixDataSourceTestClient) BaseURL() string {
	return "https://engine.test.invalid"
}

func (c networkPrefixDataSourceTestClient) Do(request *http.Request) (*http.Response, error) {
	return c.do(request)
}

func networkPrefixDataSourceTestResponse(t *testing.T, value interface{}) *http.Response {
	t.Helper()

	var body bytes.Buffer
	require.NoError(t, json.NewEncoder(&body).Encode(value))
	return &http.Response{
		Status:     "200 OK",
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewReader(body.Bytes())),
	}
}
