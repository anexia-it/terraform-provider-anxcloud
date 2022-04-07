package anxcloud

import (
	clouddns "go.anx.io/go-anxcloud/pkg/apis/clouddns/v1"
	"go.anx.io/go-anxcloud/pkg/clouddns/zone"
)

func flattenDnsZones(dnsZones []zone.Zone) []interface{} {
	zones := make([]interface{}, 0, len(dnsZones))
	if len(dnsZones) < 1 {
		return zones
	}

	for _, zone := range dnsZones {

		dnsServers := make([]interface{}, 0, len(zone.DNSServers))
		for _, dnsServer := range zone.DNSServers {
			d := map[string]interface{}{
				"server": dnsServer.Server,
				"alias":  dnsServer.Alias,
			}

			dnsServers = append(dnsServers, d)
		}

		m := map[string]interface{}{
			"name":               zone.Name,
			"is_master":          zone.IsMaster,
			"dns_sec_mode":       zone.DNSSecMode,
			"admin_email":        zone.AdminEmail,
			"refresh":            zone.Refresh,
			"retry":              zone.Retry,
			"expire":             zone.Expire,
			"ttl":                zone.TTL,
			"master_nameserver":  zone.MasterNS,
			"notify_allowed_ips": zone.NotifyAllowedIPs,
			"is_editable":        zone.IsEditable,
			"validation_level":   zone.ValidationLevel,
			"deployment_level":   zone.DeploymentLevel,
			"dns_servers":        dnsServers,
		}

		zones = append(zones, m)
	}
	return zones
}

func expandDNSServers(p []interface{}) []clouddns.DNSServer {
	dnsServers := make([]clouddns.DNSServer, 0)
	if len(p) < 1 {
		return dnsServers
	}

	for _, elem := range p {
		in := elem.(map[string]interface{})
		dnsServer := clouddns.DNSServer{}

		if v, ok := in["server"]; ok {
			dnsServer.Server = v.(string)
		}
		if v, ok := in["alias"]; ok {
			dnsServer.Alias = v.(string)
		}

		dnsServers = append(dnsServers, dnsServer)
	}

	return dnsServers
}

func flattenDNSServers(in []clouddns.DNSServer) []interface{} {
	att := []interface{}{}

	if len(in) < 1 {
		return att
	}

	for _, v := range in {
		dnsServer := map[string]interface{}{}
		dnsServer["server"] = v.Server
		dnsServer["alias"] = v.Alias
		att = append(att, dnsServer)
	}

	return att
}
