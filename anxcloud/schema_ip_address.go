package anxcloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func schemaIPAddresses() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"search": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: listQueryDescription,
		},
		"addresses": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "List of available addresses.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"identifier": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: identifierDescription,
					},
					"address": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The IP address.",
					},
					"description_customer": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Additional customer description.",
					},
					"role": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Role of the IP address.",
					},
				},
			},
		},
	}
}
