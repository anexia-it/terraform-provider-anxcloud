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
		"locations": schemaLocations(),
	}
}
