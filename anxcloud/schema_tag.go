package anxcloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func schemaTags() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"page": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     1,
			Description: "Number of page",
		},
		"limit": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     1000,
			Description: "Number of tags per page",
		},
		"query": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Search term",
		},
		"service_identifier": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The id of the service",
		},
		"organization_identifier": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The id of the organization",
		},
		"order": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The order of the tags",
		},
		"sort_ascending": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Ascending or descending",
		},
		"tags": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "List of tags.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Name of the tag.",
					},
					"identifier": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Identifier of the tag.",
					},
				},
			},
		},
	}
}

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
