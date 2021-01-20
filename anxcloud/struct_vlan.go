package anxcloud

import (
	"github.com/anexia-it/go-anxcloud/pkg/vlan"
)

// expanders

// flatteners

func flattenVLANs(in []vlan.Summary) []interface{} {
	att := []interface{}{}
	if len(in) < 1 {
		return att
	}

	for _, v := range in {
		m := map[string]interface{}{}

		m["identifier"] = v.Identifier
		m["name"] = v.Name
		m["description_customer"] = v.CustomerDescription

		att = append(att, m)
	}

	return att
}
