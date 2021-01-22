package anxcloud

import (
	"github.com/anexia-it/go-anxcloud/pkg/vsphere/provisioning/disktype"
)

// expanders

// flatteners

func flattenDiskTypes(in []disktype.DiskType) []interface{} {
	att := []interface{}{}
	if len(in) < 1 {
		return att
	}

	for _, d := range in {
		m := map[string]interface{}{}

		m["id"] = d.ID
		m["storage_type"] = d.StorageType
		m["bandwidth"] = d.Bandwidth
		m["iops"] = d.IOPS
		m["latency"] = d.Latency

		att = append(att, m)
	}

	return att
}
