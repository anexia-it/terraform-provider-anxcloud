package anxcloud

import (
	"go.anx.io/go-anxcloud/pkg/core/tags"
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

func flattenTags(in []tags.Summary) []interface{} {

	att := []interface{}{}
	if len(in) < 1 {
		return att
	}

	for _, d := range in {
		m := map[string]interface{}{}

		m["identifier"] = d.Identifier
		m["name"] = d.Name

		att = append(att, m)
	}

	return att

}

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
