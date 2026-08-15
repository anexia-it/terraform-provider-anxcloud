package anxcloud

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func schemaKubernetesNodePoolNetwork() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Node-pool network identifier.",
		},
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Network interface name.",
		},
		"node_pool": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Identifier of the node pool to which the network is attached.",
		},
		"bandwidth_limit": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Network bandwidth limit identifier, for example `1000`.",
		},
		"vlan": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Identifier of the VLAN attached to the node pool.",
		},
		"state": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Current reconciliation state identifier.",
		},
		"state_text": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Human-readable current reconciliation state.",
		},
	}
}
