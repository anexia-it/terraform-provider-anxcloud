package anxcloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func schemaVLANs() map[string]*schema.Schema {
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
		"vlans": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "List of available VLANs.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"identifier": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Identifier of the VLAN.",
					},
					"name": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "VLAN name.",
					},
					"description_customer": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Additional customer description.",
					},
				},
			},
		},
	}
}

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
