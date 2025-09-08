package anxcloud

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"go.anx.io/go-anxcloud/pkg/api"
	"go.anx.io/go-anxcloud/pkg/api/types"

	objectstoragev2 "go.anx.io/go-anxcloud/pkg/apis/objectstorage/v2"
)

func dataSourceObjectStorageEndpoints() *schema.Resource {
	return &schema.Resource{
		Description: "Provides a list of Object Storage endpoints.",
		ReadContext: dataSourceObjectStorageEndpointsRead,
		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(1 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"page": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1,
				Description: "Page number for pagination.",
			},
			"limit": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     50,
				Description: "Number of items per page.",
			},
			"search": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Search term to filter endpoints.",
			},
			"enabled_only": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "If true, only return enabled endpoints.",
			},
			"endpoints": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of endpoints.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"identifier": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The identifier of the endpoint.",
						},
						"url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "URL of the endpoint.",
						},
						"endpoint_user": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the endpoint user.",
						},
						"enabled": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Indicates if the endpoint is enabled.",
						},
						"state": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "State of the endpoint.",
						},
						"reseller": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Reseller identifier.",
						},
						"customer": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Customer identifier.",
						},
						"created_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The creation time of the endpoint.",
						},
						"updated_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The last update time of the endpoint.",
						},
					},
				},
			},
		},
	}
}

func dataSourceObjectStorageEndpointsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	a := apiFromProviderConfig(m)

	// Get configuration
	enabledOnly := d.Get("enabled_only").(bool)

	// Get list of endpoints without any filtering first
	var pageIter types.PageInfo
	err := a.List(ctx, &objectstoragev2.Endpoint{}, api.Paged(1, 100, &pageIter))
	if err != nil {
		return diag.FromErr(err)
	}

	endpoints := make([]objectstoragev2.Endpoint, 0, pageIter.TotalItems())
	var pagedEndpoints []objectstoragev2.Endpoint
	for pageIter.Next(&pagedEndpoints) {
		endpoints = append(endpoints, pagedEndpoints...)
	}

	if err := pageIter.Error(); err != nil {
		return diag.FromErr(err)
	}

	// Filter endpoints based on enabled status if requested
	var filteredEndpoints []objectstoragev2.Endpoint
	if enabledOnly {
		for _, endpoint := range endpoints {
			if endpoint.Enabled {
				filteredEndpoints = append(filteredEndpoints, endpoint)
			}
		}
	} else {
		filteredEndpoints = endpoints
	}

	// Convert to terraform data structure
	endpointList := make([]interface{}, len(filteredEndpoints))
	for i, endpoint := range filteredEndpoints {
		endpointMap := map[string]interface{}{
			"identifier":    endpoint.Identifier,
			"url":           endpoint.URL,
			"endpoint_user": endpoint.EndpointUser,
			"enabled":       endpoint.Enabled,
		}

		// Handle state
		if endpoint.State != nil {
			endpointMap["state"] = *endpoint.State
		}

		// Handle organization references
		if endpoint.Reseller != "" {
			endpointMap["reseller"] = endpoint.Reseller
		}
		if endpoint.Customer != "" {
			endpointMap["customer"] = endpoint.Customer
		}

		// Note: CreatedAt and UpdatedAt fields are not available in the current API structure

		endpointList[i] = endpointMap
	}

	if err := d.Set("endpoints", endpointList); err != nil {
		return diag.FromErr(err)
	}

	// Set ID based on time to make resource unique
	d.SetId(generateDataSourceID())

	return diags
}
