package anxcloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func schemaLocations() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Description: "Anexia Cloud Locations.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"identifier": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Identifier of the location.",
				},
				"name": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Name of the location.",
				},
				"code": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Location code.",
				},
				"country": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Location country.",
				},
				"city_code": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Location city code.",
				},
				"lat": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Location latitude.",
				},
				"lon": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Location longitude.",
				},
			},
		},
	}
}
