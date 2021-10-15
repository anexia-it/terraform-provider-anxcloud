package anxcloud

import (
	"github.com/anexia-it/go-anxcloud/pkg/clouddns/zone"
	"github.com/google/go-cmp/cmp"
	"testing"
)

func TestFlattenDnsZones(t *testing.T) {
	cases := []struct {
		Input          []zone.Zone
		ExpectedOutput []interface{}
	}{
		{
			[]zone.Zone{
				{
					Definition: &zone.Definition{
						Name:             "zone1.test",
						IsMaster:         true,
						DNSSecMode:       "unvalidated",
						AdminEmail:       "test@zone1.test",
						Refresh:          3600,
						Retry:            300,
						Expire:           3600,
						TTL:              60,
						NotifyAllowedIPs: []string{"127.0.0.1", "192.168.0.1"},
						MasterNS:         "8.8.8.8",
						DNSServers: []zone.DNSServer{
							{
								Server: "nameserver-1",
								Alias:  "ns1",
							},
						},
					},
					DeploymentLevel: 100,
					ValidationLevel: 100,
					IsEditable:      true,
				},
				{
					Definition: &zone.Definition{
						Name:             "zone2.test",
						IsMaster:         true,
						DNSSecMode:       "managed",
						AdminEmail:       "test@zone2.test",
						Refresh:          3600,
						Retry:            300,
						Expire:           3600,
						TTL:              60,
						NotifyAllowedIPs: []string{"127.0.0.1", "192.168.0.1"},
						MasterNS:         "8.8.8.8",
						DNSServers: []zone.DNSServer{
							{
								Server: "nameserver-2",
								Alias:  "ns2",
							},
						},
					},
					DeploymentLevel: 99,
					ValidationLevel: 99,
					IsEditable:      false,
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
			[]zone.Zone{},
			[]interface{}{},
		},
	}

	for _, tc := range cases {
		output := flattenDnsZones(tc.Input)
		if diff := cmp.Diff(tc.ExpectedOutput, output); diff != "" {
			t.Fatalf("Unexpected output from flattener: missmatch (-want +got):\n%s", diff)
		}
	}
}
