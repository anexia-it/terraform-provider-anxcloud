package anxcloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func schemaCPUPerformanceType() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"types": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "List of CPU performance types.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"id": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "CPU performance type identifier.",
					},
					"prioritization": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "CPU performance type prioritization.",
					},
					"limit": {
						Type:        schema.TypeFloat,
						Computed:    true,
						Description: "CPU performance type limit.",
					},
					"unit": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "CPU performance type limit unit.",
					},
				},
			},
		},
	}
}
