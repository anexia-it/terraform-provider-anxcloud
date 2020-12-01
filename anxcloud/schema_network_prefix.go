package anxcloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func schemaNetworkPrefix() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"location_id": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "Location Identifier.",
		},
		"netmask": {
			Type:        schema.TypeInt,
			Required:    true,
			ForceNew:    true,
			Description: "Netmask size. Example: 29 which would result in x.x.x.x/29.",
		},
		"ip_version": {
			Type:        schema.TypeInt,
			Optional:    true,
			ForceNew:    true,
			Description: "The Prefix version: 4 = IPv4, 6 = IPv6.",
		},
		"type": {
			Type:        schema.TypeInt,
			Optional:    true,
			ForceNew:    true,
			Description: "The Prefix type: 0 = Public, 1 = Private.",
		},
		"new_vlan": {
			Type:     schema.TypeBool,
			Optional: true,
			ForceNew: true,
			Default:  false,
			Description: "If new VLAN shall be created. WARNING, the VLAN status won't be reflected in the terraform status." +
				"Use at your own risk.",
		},
		"vlan_id": {
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
			Description: "Identifier for the related VLAN.",
		},
		"router_redundancy": {
			Type:        schema.TypeBool,
			Optional:    true,
			ForceNew:    true,
			Default:     false,
			Description: "If router Redundancy shall be enabled.",
		},
		"vm_provisioning": {
			Type:        schema.TypeBool,
			Optional:    true,
			ForceNew:    true,
			Default:     false,
			Description: "If VM provisioning shall be enabled.",
		},
		"description_customer": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Additional description.",
		},
		"description_vlan_customer": {
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
			Description: "Additional description for the generated VLAN if new_vlan.",
		},
		"organization": {
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
			Description: "Customer of yours. Reseller only.",
		},
		"status": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Status of the created prefix.",
		},
		"cidr": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "CIDR of the created prefix.",
		},
		"role_text": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Role of the prefix.",
		},
		"description_internal": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Internal description.",
		},
		"locations": schemaLocations(),
	}
}
