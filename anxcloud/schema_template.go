package anxcloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func schemaTemplate() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"location_id": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Location identifier.",
		},
		"template_type": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "templates",
			Description: "Template type. Defaults to 'templates'.",
		},
		"templates": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "Available list of templates.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"id": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Template identifier.",
					},
					"name": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "OS name.",
					},
					"bit": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "OS bit.",
					},
					"build": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "OS build.",
					},
				},
			},
		},
	}
}
