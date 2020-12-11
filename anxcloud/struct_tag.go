package anxcloud

import (
	"github.com/anexia-it/go-anxcloud/pkg/core/tags"
)

// expanders

func expandTags(p []interface{}) []string {
	var out []string
	for _, elem := range p {
		out = append(out, elem.(string))
	}
	return out
}

// flatteners

func flattenOrganisationAssignments(in []tags.Organisation) []interface{} {
	att := []interface{}{}
	if len(in) < 1 {
		return att
	}

	for _, v := range in {
		m := map[string]interface{}{}

		m["customer"] = []map[string]interface{}{
			{
				"id":          v.Customer.Identifier,
				"customer_id": v.Customer.CustomerID,
				"demo":        v.Customer.Demo,
				"name":        v.Customer.Name,
				"name_slug":   v.Customer.Slug,
				"reseller":    v.Customer.Reseller,
			},
		}

		m["service"] = []map[string]interface{}{
			{
				"id":   v.Service.Identifier,
				"name": v.Service.Name,
			},
		}

		att = append(att, m)
	}

	return att
}
