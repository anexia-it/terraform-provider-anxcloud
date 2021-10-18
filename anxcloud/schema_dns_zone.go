package anxcloud

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func schemaDNSZones() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"zones": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "List of DNS zones of the customer",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The zone name",
					},
					"is_master": {
						Type:        schema.TypeBool,
						Computed:    true,
						Description: "Indicator if the zone is a master zone",
					},
					"dns_sec_mode": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: `DNSSec mode value for master zones. ["managed" or "unvalidated"]`,
					},
					"admin_email": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Admin email address",
					},
					"refresh": {
						Type:        schema.TypeInt,
						Computed:    true,
						Description: "Refresh value",
					},
					"retry": {
						Type:        schema.TypeInt,
						Computed:    true,
						Description: "Retry value",
					},
					"expire": {
						Type:        schema.TypeInt,
						Computed:    true,
						Description: "Expiration value",
					},
					"ttl": {
						Type:        schema.TypeInt,
						Computed:    true,
						Description: "TTL value of a zone",
					},
					"master_nameserver": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Master nameserver",
					},
					"notify_allowed_ips": {
						Type:        schema.TypeList,
						Computed:    true,
						Description: "IP addresses allowed to initiate domain transfer",
						Elem:        schema.TypeString,
					},
					"dns_servers": {
						Type:        schema.TypeList,
						Computed:    true,
						Description: "Configured DNS servers",
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"server": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "DNS server name",
								},
								"alias": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "DNS server alias",
								},
							},
						},
					},
					"is_editable": {
						Type:        schema.TypeBool,
						Computed:    true,
						Description: "Indicator if zone is editable",
					},
					"validation_level": {
						Type:        schema.TypeInt,
						Computed:    true,
						Description: "Current validation level in percent",
					},
					"deployment_level": {
						Type:        schema.TypeInt,
						Computed:    true,
						Description: "Current state of deployment progress in percent",
					},
				},
			},
		},
	}
}
