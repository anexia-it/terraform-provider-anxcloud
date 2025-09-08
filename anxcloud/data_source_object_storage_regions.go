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

func dataSourceObjectStorageRegions() *schema.Resource {
	return &schema.Resource{
		Description: "Provides a list of Object Storage regions.",
		ReadContext: dataSourceObjectStorageRegionsRead,
		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(1 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"name_filter": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter regions by name (partial match).",
			},
			"backend_filter": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter regions by backend identifier.",
			},
			"regions": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of regions.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"identifier": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The identifier of the region.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the region.",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Description of the region.",
						},
						"backend": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Identifier of the S3 backend this region belongs to.",
						},
						"state": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "State of the region.",
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
					},
				},
			},
		},
	}
}

func dataSourceObjectStorageRegionsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	a := apiFromProviderConfig(m)

	// Get configuration
	nameFilter := d.Get("name_filter").(string)
	backendFilter := d.Get("backend_filter").(string)

	// Get list of regions
	var pageIter types.PageInfo
	err := a.List(ctx, &objectstoragev2.Region{}, api.Paged(1, 100, &pageIter))
	if err != nil {
		return diag.FromErr(err)
	}

	regions := make([]objectstoragev2.Region, 0, pageIter.TotalItems())
	var pagedRegions []objectstoragev2.Region
	for pageIter.Next(&pagedRegions) {
		regions = append(regions, pagedRegions...)
	}

	if err := pageIter.Error(); err != nil {
		return diag.FromErr(err)
	}

	// Filter regions
	var filteredRegions []objectstoragev2.Region
	for _, region := range regions {
		// Filter by backend if provided
		if backendFilter != "" && (region.Backend == nil || region.Backend.Identifier != backendFilter) {
			continue
		}

		// Filter by name if provided
		if nameFilter != "" && !contains(region.Name, nameFilter) {
			continue
		}

		filteredRegions = append(filteredRegions, region)
	}

	// Convert to terraform data structure
	regionList := make([]interface{}, len(filteredRegions))
	for i, region := range filteredRegions {
		regionMap := map[string]interface{}{
			"identifier":  region.Identifier,
			"name":        region.Name,
			"description": region.Description,
		}

		// Handle backend reference
		if region.Backend != nil {
			regionMap["backend"] = region.Backend.Identifier
		} else {
			regionMap["backend"] = ""
		}

		// Handle optional fields
		if region.State != nil {
			regionMap["state"] = region.State.String()
		}

		// Handle organization references
		if region.Reseller != "" {
			regionMap["reseller"] = region.Reseller
		}
		if region.Customer != "" {
			regionMap["customer"] = region.Customer
		}

		// Note: CreatedAt and UpdatedAt fields are not available in the current API structure

		regionList[i] = regionMap
	}

	if err := d.Set("regions", regionList); err != nil {
		return diag.FromErr(err)
	}

	// Set ID based on time to make resource unique
	d.SetId(generateDataSourceID())

	return diags
}
