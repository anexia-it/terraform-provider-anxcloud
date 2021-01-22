package anxcloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func schemaDataSourceLocation() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"page": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     1,
			Description: "The page of the list.",
		},
		"limit": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     1000,
			Description: "The records limit of the list.",
		},
		"search": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "An optional string allowing to search trough entities.",
		},
		"locations": schemaLocations(),
	}
}

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
