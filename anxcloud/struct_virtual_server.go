package anxcloud

import (
	"github.com/anexia-it/go-anxcloud/pkg/vsphere/info"
	"github.com/anexia-it/go-anxcloud/pkg/vsphere/provisioning/vm"
)

// expanders

func expandVirtualServerNetworks(p []interface{}) []vm.Network {
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

func expandVirtualServerDNS(p []interface{}) (dns [maxDNSEntries]string) {
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

func flattenVirtualServerInfo(in *info.Info) []interface{} {
	if in == nil {
		return []interface{}{}
	}

	att := map[string]interface{}{}
	att["status"] = in.Status
	att["name"] = in.Name
	att["custom_name"] = in.CustomName
	att["location_code"] = in.LocationCode
	att["location_country"] = in.LocationCountry
	att["location_name"] = in.LocationName
	att["disks_number"] = in.Disks
	att["guest_os"] = in.GuestOS
	att["version_tools"] = in.VersionTools
	att["guest_tools_status"] = in.GuestToolsStatus

	var disksInfo []map[string]interface{}
	for _, d := range in.DiskInfo {
		di := map[string]interface{}{}
		di["disk_id"] = d.DiskID
		di["disk_gb"] = d.DiskGB
		di["disk_type"] = d.DiskType
		di["iops"] = d.IOPS
		di["latency"] = d.Latency
		di["storage_type"] = d.StorageType
		di["bus_type"] = d.BusType
		di["bus_type_label"] = d.BusTypeLabel
		disksInfo = append(disksInfo, di)
	}
	att["disks_info"] = disksInfo

	var networkInfo []map[string]interface{}
	for _, n := range in.Network {
		ni := map[string]interface{}{}
		ni["id"] = n.ID
		ni["nic"] = n.NIC
		ni["vlan"] = n.VLAN
		ni["mac_address"] = n.MACAddress
		ni["ip_v4"] = n.IPv4
		ni["ip_v6"] = n.IPv6
		networkInfo = append(networkInfo, ni)
	}
	att["network"] = networkInfo

	return []interface{}{att}
}
