package anxcloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func schemaDiskTypes() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"location_id": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Location identifier.",
		},
		"types": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "List of available disk types.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"id": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: identifierDescription,
					},
					"storage_type": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Storage type.",
					},
					"bandwidth": {
						Type:        schema.TypeInt,
						Computed:    true,
						Description: "Bandwidth.",
					},
					"iops": {
						Type:        schema.TypeInt,
						Computed:    true,
						Description: "Input/output operations per second.",
					},
					"latency": {
						Type:        schema.TypeInt,
						Computed:    true,
						Description: "Latency.",
					},
				},
			},
		},
	}
}
