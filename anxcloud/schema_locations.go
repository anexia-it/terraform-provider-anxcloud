package anxcloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func schemaDataSourceLocations() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"search": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: listQueryDescription,
		},
		"locations": schemaLocations(),
	}
}

func schemaLocation() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"identifier": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: identifierDescription,
		},
		"name": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Location name.",
		},
		"code": {
			Type:        schema.TypeString,
			Optional:    true,
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
	}
}

func schemaLocations() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Description: "Anexia Cloud Locations.",
		Elem: &schema.Resource{
			Schema: schemaLocation(),
		},
	}
}
