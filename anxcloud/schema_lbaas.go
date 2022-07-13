package anxcloud

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func schemaLBaaSLoadBalancer() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "LoadBalancer name.",
		},
		"ip_address": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Address of the Engine management API of the LoadBalancer.",
		},
	}
}
