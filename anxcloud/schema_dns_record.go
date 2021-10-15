package anxcloud

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func schemaDNSRecords() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"zone_name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The parent zone",
		},
		"records": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "List of known records on the zone",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"identifier": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The record's identifier",
					},
					"type": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The DNS record type",
					},
					"rdata": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The DNS record data",
					},
					"name": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The name of the DNS record",
					},
					"region": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The region for geodns aware records",
					},
					"ttl": {
						Type:        schema.TypeInt,
						Computed:    true,
<<<<<<< HEAD
						Description: "The region specific TTL. If nil the zone TTL will be used",
=======
						Description: "The region specific TTL. If null the zone TTL will be used",
>>>>>>> Added datasource for DNS records
					},
					"immutable": {
						Type:        schema.TypeBool,
						Computed:    true,
						Description: "Specifies wheather or not a record is immutable",
					},
				},
			},
		},
	}
}
