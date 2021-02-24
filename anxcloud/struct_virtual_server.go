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

func expandVirtualServerDisks(p []interface{}) []vm.Disk {
	var disks []vm.Disk
	if len(p) < 1 {
		return disks
	}

	for _, elem := range p {
		in := elem.(map[string]interface{})
		disk := vm.Disk{}

		if v, ok := in["disk_type"]; ok {
			disk.Type = v.(string)
		}
		if v, ok := in["disk"]; ok {
			disk.SizeGBs = v.(int)
		}
		//if v, ok := in["disk_id"]; ok {
		//	disk.ID = v.(int)
		//}

		disks = append(disks, disk)
	}

	return disks
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

func expandVirtualServerInfo(p []interface{}) info.Info {
	var i info.Info
	if len(p) < 1 {
		return i
	}

	att := p[0].(map[string]interface{})
	if v, ok := att["identifier"]; ok {
		i.Identifier = v.(string)
	}
	if v, ok := att["status"]; ok {
		i.Status = v.(string)
	}
	if v, ok := att["name"]; ok {
		i.Name = v.(string)
	}
	if v, ok := att["custom_name"]; ok {
		i.CustomName = v.(string)
	}
	if v, ok := att["location_code"]; ok {
		i.LocationCode = v.(string)
	}
	if v, ok := att["location_country"]; ok {
		i.LocationCountry = v.(string)
	}
	if v, ok := att["location_name"]; ok {
		i.LocationName = v.(string)
	}
	if v, ok := att["cpu"]; ok {
		i.CPU = v.(int)
	}
	if v, ok := att["cores"]; ok {
		i.Cores = v.(int)
	}
	if v, ok := att["ram"]; ok {
		i.RAM = v.(int)
	}
	if v, ok := att["disks_number"]; ok {
		i.Disks = v.(int)
	}
	if v, ok := att["guest_os"]; ok {
		i.GuestOS = v.(string)
	}
	if v, ok := att["version_tools"]; ok {
		i.VersionTools = v.(string)
	}
	if v, ok := att["guest_tools_status"]; ok {
		i.GuestToolsStatus = v.(string)
	}

	if v, ok := att["disks_info"]; ok {
		disks := v.([]interface{})

		for _, elem := range disks {
			disk := info.DiskInfo{}
			d := elem.(map[string]interface{})

			if v, ok := d["disk_id"]; ok {
				disk.DiskID = v.(int)
			}
			if v, ok := d["disk_gb"]; ok {
				disk.DiskGB = v.(int)
			}
			if v, ok := d["disk_type"]; ok {
				disk.DiskType = v.(string)
			}
			if v, ok := d["iops"]; ok {
				disk.IOPS = v.(int)
			}
			if v, ok := d["latency"]; ok {
				disk.Latency = v.(int)
			}
			if v, ok := d["storage_type"]; ok {
				disk.StorageType = v.(string)
			}
			if v, ok := d["bus_type"]; ok {
				disk.BusType = v.(string)
			}
			if v, ok := d["bus_type_label"]; ok {
				disk.BusTypeLabel = v.(string)
			}

			i.DiskInfo = append(i.DiskInfo, disk)
		}
	}

	if v, ok := att["network"]; ok {
		networks := v.([]interface{})

		for _, elem := range networks {
			network := info.Network{}
			n := elem.(map[string]interface{})

			if v, ok := n["id"]; ok {
				network.ID = v.(int)
			}
			if v, ok := n["nic"]; ok {
				network.NIC = v.(int)
			}
			if v, ok := n["vlan"]; ok {
				network.VLAN = v.(string)
			}
			if v, ok := n["mac_address"]; ok {
				network.MACAddress = v.(string)
			}
			if v, ok := n["ip_v4"]; ok {
				for _, ip := range v.([]interface{}) {
					network.IPv4 = append(network.IPv4, ip.(string))
				}
			}
			if v, ok := n["ip_v6"]; ok {
				for _, ip := range v.([]interface{}) {
					network.IPv6 = append(network.IPv6, ip.(string))
				}
			}

			i.Network = append(i.Network, network)
		}
	}

	return i
}

// flatteners

func flattenVirtualServerNetwork(in []vm.Network) []interface{} {
	att := []interface{}{}
	if len(in) < 1 {
		return att
	}

	for _, n := range in {
		net := map[string]interface{}{}
		net["vlan_id"] = n.VLAN
		net["nic_type"] = n.NICType
		net["ips"] = n.IPs
		att = append(att, net)
	}

	return att
}

func flattenVirtualServerInfo(in *info.Info) []interface{} {
	if in == nil {
		return []interface{}{}
	}

	att := map[string]interface{}{}
	att["identifier"] = in.Identifier
	att["status"] = in.Status
	att["name"] = in.Name
	att["custom_name"] = in.CustomName
	att["location_code"] = in.LocationCode
	att["location_country"] = in.LocationCountry
	att["location_name"] = in.LocationName
	att["cpu"] = in.CPU
	att["cores"] = in.Cores
	att["ram"] = in.RAM
	att["disks_number"] = in.Disks
	att["guest_os"] = in.GuestOS
	att["version_tools"] = in.VersionTools
	att["guest_tools_status"] = in.GuestToolsStatus

	disksInfo := []interface{}{}
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

	networkInfo := []interface{}{}
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

func flattenVirtualServerDisks(in []vm.Disk) []interface{} {
	att := []interface{}{}
	if len(in) < 1 {
		return att
	}

	for _, d := range in {
		net := map[string]interface{}{}
		net["disk_type"] = d.Type
		net["disk"] = d.SizeGBs
		//net["disk_id"] = d.ID
		att = append(att, net)
	}

	return att
}
