package anxcloud

import (
	"github.com/anexia-it/go-anxcloud/pkg/vsphere/provisioning/vm"
)

// expanders

func expandNetworks(p []interface{}) []vm.Network {
	var networks []vm.Network
	if len(p) < 1 {
		return networks
	}

	for _, elem := range p {
		in := elem.(map[string]interface{})
		network := vm.Network{}

		if v, ok := in["vlan_id"]; ok {
			network.VLAN = v.(string)
		}
		if v, ok := in["nic_type"]; ok {
			network.NICType = v.(string)
		}
		if v, ok := in["ips"]; ok {
			ips := v.([]interface{})
			for _, ip := range ips {
				network.IPs = append(network.IPs, ip.(string))
			}
		}

		networks = append(networks, network)
	}

	return networks
}

func expandDNS(p []interface{}) (dns [maxDNSLen]string) {
	if len(p) < 1 {
		return dns
	}

	for i, elem := range p {
		if i > len(dns) {
			return dns
		}
		dns[i] = elem.(string)
	}

	return dns
}

// flatteners
