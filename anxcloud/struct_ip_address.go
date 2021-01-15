package anxcloud

import (
	"github.com/anexia-it/go-anxcloud/pkg/ipam/address"
)

// expanders

// flatteners

func flattenIPAddresses(in []address.Summary) []interface{} {
	att := []interface{}{}
	if len(in) < 1 {
		return att
	}

	for _, a := range in {
		m := map[string]interface{}{}

		m["identifier"] = a.ID
		m["address"] = a.Name
		m["description_customer"] = a.DescriptionCustomer
		m["role"] = a.Role

		att = append(att, m)
	}

	return att
}
