package anxcloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func schemaNICTypes() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"nic_types": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "List of available nic types.",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
	}
}
