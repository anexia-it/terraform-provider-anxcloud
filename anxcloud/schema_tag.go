package anxcloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func schemaTag() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "The tag name.",
		},
		"service_id": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "The identifier of the service this tag should be assigned to.",
		},
		"customer_id": {
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
			Description: "The identifier of the customer this tag should be assigned to. Leave empty to assign to the organization of the logged in user.",
		},
		"organisation_assignments": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "Organisation assignments.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"customer": {
						Type:        schema.TypeList,
						Computed:    true,
						Description: "Customer related info.",
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"id": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "Identifier.",
								},
								"customer_id": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "Identifier of the customer.",
								},
								"demo": {
									Type:        schema.TypeBool,
									Computed:    true,
									Description: "Whether is demo.",
								},
								"name": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "Customer name.",
								},
								"name_slug": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "Slug name.",
								},
								"reseller": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "Reseller.",
								},
							},
						},
					},
					"service": {
						Type:        schema.TypeList,
						Computed:    true,
						Description: "Service related info.",
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"name": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "Name of the service.",
								},
								"id": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "Identifier of the service.",
								},
							},
						},
					},
				},
			},
		},
	}
}
