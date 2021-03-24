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
		"vlan_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The corresponding VLAN ID",
		},
		"router_redundancy": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "If router Redundancy shall be enabled.",
		},
		"description_customer": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Additional description.",
		},
		"organization": {
			Type:        schema.TypeString,
			Optional:    true,
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
