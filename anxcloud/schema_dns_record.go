package anxcloud

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func schemaDNSRecord() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"identifier": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "DNS Record identifier. Changes on revision change and therefore shouldn't be used as reference.",
		},
		"type": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "DNS record type.",
			ForceNew:    true,
		},
		"rdata": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "DNS record data.",
			ForceNew:    true,
		},
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "DNS record name.",
			ForceNew:    true,
		},
		"zone_name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Zone of DNS record.",
			ForceNew:    true,
		},
		"ttl": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "Region specific TTL. If not set the zone TTL will be used.",
			ForceNew:    true,
		},

		"region": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "DNS record region (for GeoDNS aware records).",
		},
		"immutable": {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: "Specifies whether or not a record is immutable.",
		},
	}
}

func schemaDNSRecords() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"zone_name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Parent zone",
		},
		"records": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "List of known records on the zone",
			Elem: &schema.Resource{
				Schema: schemaDNSRecord(),
			},
		},
	}
}
