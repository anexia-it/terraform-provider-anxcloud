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
						Description: "Id of the CPU performance type",
					},
					"prioritization": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Prio of the CPU performance type",
					},
					"limit": {
						Type:        schema.TypeFloat,
						Computed:    true,
						Description: "The limit of the CPU performance type",
					},
					"unit": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The unit for the limit of the CPU performance type",
					},
				},
			},
		},
	}
}
