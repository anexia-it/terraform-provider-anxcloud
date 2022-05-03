package anxcloud

import (
	clouddnsv1 "go.anx.io/go-anxcloud/pkg/apis/clouddns/v1"
)

func flattenDnsZones(dnsZones []clouddnsv1.Zone) []interface{} {
	zones := make([]interface{}, 0, len(dnsZones))
	if len(dnsZones) < 1 {
		return zones
	}

	for _, zone := range dnsZones {
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
			"dns_servers":        flattenDNSServers(zone.DNSServers),
		}

		zones = append(zones, m)
	}
	return zones
}

func expandDNSServers(p []interface{}) []clouddnsv1.DNSServer {
	dnsServers := make([]clouddnsv1.DNSServer, 0, len(p))
	if len(p) < 1 {
		return dnsServers
	}

	for _, elem := range p {
		in := elem.(map[string]interface{})
		dnsServer := clouddnsv1.DNSServer{}

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

func flattenDNSServers(in []clouddnsv1.DNSServer) []interface{} {
	att := make([]interface{}, 0, len(in))

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
