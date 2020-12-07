package anxcloud

import (
	"github.com/anexia-it/go-anxcloud/pkg/vsphere/provisioning/templates"
)

// expanders

// flatteners

func flattenTemplates(in []templates.Template) []interface{} {
	att := []interface{}{}
	if len(in) < 1 {
		return att
	}

	for _, t := range in {
		m := map[string]interface{}{}

		m["id"] = t.ID
		m["name"] = t.Name
		m["build"] = t.Build
		m["bit"] = t.WordSize

		att = append(att, m)
	}

	return att
}
