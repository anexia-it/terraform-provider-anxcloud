package anxcloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func schemaVLAN() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"location_id": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "ANX Location Identifier.",
		},
		"vm_provisioning": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			ForceNew:    true,
			Description: "True if VM provisioning shall be enabled. Defaults to false.",
		},
		"description_customer": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Additional customer description.",
		},
		"name": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "VLAN name.",
		},
		"description_internal": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Internal description.",
		},
		"role_text": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Role of the VLAN.",
		},
		"status": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "VLAN status.",
		},
		"locations": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "Role of the VLAN.",
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
		},
	}
}
