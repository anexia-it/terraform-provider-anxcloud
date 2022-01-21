package anxcloud

import (
	cpuperformancetype "go.anx.io/go-anxcloud/pkg/vsphere/provisioning/cpuperformancetypes"
)

// expanders

// flatteners

func flattenCPUPerformanceTypes(in []cpuperformancetype.CPUPerformanceType) []interface{} {

	att := []interface{}{}
	if len(in) < 1 {
		return att
	}

	for _, d := range in {
		m := map[string]interface{}{}

		m["id"] = d.ID
		m["prioritization"] = d.Prioritization
		m["limit"] = d.Limit
		m["unit"] = d.Unit

		att = append(att, m)
	}

	return att

}
