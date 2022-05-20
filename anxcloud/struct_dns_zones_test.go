package anxcloud

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	clouddnsv1 "go.anx.io/go-anxcloud/pkg/apis/clouddns/v1"
)

func TestFlattenDnsZones(t *testing.T) {
	cases := []struct {
		Input          []clouddnsv1.Zone
		ExpectedOutput []interface{}
	}{
		{
			[]clouddnsv1.Zone{
				{
					Name:             "zone1.test",
					IsMaster:         true,
					IsEditable:       true,
					DNSSecMode:       "unvalidated",
					AdminEmail:       "test@zone1.test",
					Refresh:          3600,
					Retry:            300,
					Expire:           3600,
					TTL:              60,
					NotifyAllowedIPs: []string{"127.0.0.1", "192.168.0.1"},
					MasterNS:         "8.8.8.8",
					DeploymentLevel:  100,
					ValidationLevel:  100,
					DNSServers: []clouddnsv1.DNSServer{
						{
							Server: "nameserver-1",
							Alias:  "ns1",
						},
					},
				},
				{
					Name:             "zone2.test",
					IsMaster:         true,
					IsEditable:       false,
					DNSSecMode:       "managed",
					AdminEmail:       "test@zone2.test",
					Refresh:          3600,
					Retry:            300,
					Expire:           3600,
					TTL:              60,
					NotifyAllowedIPs: []string{"127.0.0.1", "192.168.0.1"},
					MasterNS:         "8.8.8.8",
					DeploymentLevel:  99,
					ValidationLevel:  99,
					DNSServers: []clouddnsv1.DNSServer{
						{
							Server: "nameserver-2",
							Alias:  "ns2",
						},
					},
				},
			},
			[]interface{}{
				map[string]interface{}{
					"name":               "zone1.test",
					"is_master":          true,
					"dns_sec_mode":       "unvalidated",
					"admin_email":        "test@zone1.test",
					"refresh":            3600,
					"retry":              300,
					"expire":             3600,
					"ttl":                60,
					"notify_allowed_ips": []string{"127.0.0.1", "192.168.0.1"},
					"master_nameserver":  "8.8.8.8",
					"deployment_level":   100,
					"validation_level":   100,
					"is_editable":        true,
					"dns_servers": []interface{}{
						map[string]interface{}{
							"server": "nameserver-1",
							"alias":  "ns1",
						},
					},
				},
				map[string]interface{}{
					"name":               "zone2.test",
					"is_master":          true,
					"dns_sec_mode":       "managed",
					"admin_email":        "test@zone2.test",
					"refresh":            3600,
					"retry":              300,
					"expire":             3600,
					"ttl":                60,
					"notify_allowed_ips": []string{"127.0.0.1", "192.168.0.1"},
					"master_nameserver":  "8.8.8.8",
					"deployment_level":   99,
					"validation_level":   99,
					"is_editable":        false,
					"dns_servers": []interface{}{
						map[string]interface{}{
							"server": "nameserver-2",
							"alias":  "ns2",
						},
					},
				},
			},
		},
		{
			[]clouddnsv1.Zone{},
			[]interface{}{},
		},
	}

	for _, tc := range cases {
		output := flattenDNSZones(tc.Input)
		if diff := cmp.Diff(tc.ExpectedOutput, output); diff != "" {
			t.Fatalf("Unexpected output from flattener: missmatch (-want +got):\n%s", diff)
		}
	}
}

func TestFlattenDNSServers(t *testing.T) {
	cases := []struct {
		Input          []clouddnsv1.DNSServer
		ExpectedOutput []interface{}
	}{
		{
			[]clouddnsv1.DNSServer{
				{
					Server: "ns1.example.com",
					Alias:  "Nameserver #1",
				},
				{
					Server: "ns2.example.com",
				},
			},
			[]interface{}{
				map[string]interface{}{
					"server": "ns1.example.com",
					"alias":  "Nameserver #1",
				},
				map[string]interface{}{
					"server": "ns2.example.com",
					"alias":  "",
				},
			},
		},
		{
			[]clouddnsv1.DNSServer{},
			[]interface{}{},
		},
	}

	for _, tc := range cases {
		output := flattenDNSServers(tc.Input)
		if diff := cmp.Diff(tc.ExpectedOutput, output); diff != "" {
			t.Fatalf("Unexpected output from flattener: missmatch (-want +got):\n%s", diff)
		}
	}
}

func TestExpandDNSServers(t *testing.T) {
	cases := []struct {
		Input          []interface{}
		ExpectedOutput []clouddnsv1.DNSServer
	}{
		{
			[]interface{}{
				map[string]interface{}{
					"server": "ns1.example.com",
					"alias":  "Nameserver #1",
				},
				map[string]interface{}{
					"server": "ns2.example.com",
				},
			},
			[]clouddnsv1.DNSServer{
				{
					Server: "ns1.example.com",
					Alias:  "Nameserver #1",
				},
				{
					Server: "ns2.example.com",
				},
			},
		},
		{
			[]interface{}{},
			[]clouddnsv1.DNSServer{},
		},
	}

	for _, tc := range cases {
		output := expandDNSServers(tc.Input)
		if diff := cmp.Diff(tc.ExpectedOutput, output); diff != "" {
			t.Fatalf("Unexpected output from expander: missmatch (-want +got):\n%s", diff)
		}
	}
}
