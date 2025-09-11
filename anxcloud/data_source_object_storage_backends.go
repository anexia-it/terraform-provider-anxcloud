package anxcloud

import (
	"context"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"go.anx.io/go-anxcloud/pkg/api"
	"go.anx.io/go-anxcloud/pkg/api/types"

	objectstoragev2 "go.anx.io/go-anxcloud/pkg/apis/objectstorage/v2"
)

func dataSourceObjectStorageBackends() *schema.Resource {
	return &schema.Resource{
		Description: "Provides a list of Object Storage S3 backends.",
		ReadContext: dataSourceObjectStorageBackendsRead,
		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(1 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"name_filter": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter backends by name (partial match).",
			},
			"enabled_only": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "If true, only return enabled backends.",
			},
			"backends": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of backends.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"identifier": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The identifier of the backend.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the S3 backend.",
						},
						"endpoint": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Endpoint identifier this backend belongs to.",
						},
						"backend_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Type of the S3 backend.",
						},
						"enabled": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Indicates if the S3 backend is enabled.",
						},
						"backend_user": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the S3 backend user.",
						},
						"state": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "State of the backend.",
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

func dataSourceObjectStorageBackendsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	a := apiFromProviderConfig(m)

	// Get configuration
	nameFilter := d.Get("name_filter").(string)
	enabledOnly := d.Get("enabled_only").(bool)

	// Get list of backends
	var pageIter types.PageInfo
	err := a.List(ctx, &objectstoragev2.S3Backend{}, api.Paged(1, 100, &pageIter))
	if err != nil {
		return diag.FromErr(err)
	}

	backends := make([]objectstoragev2.S3Backend, 0, pageIter.TotalItems())
	var pagedBackends []objectstoragev2.S3Backend
	for pageIter.Next(&pagedBackends) {
		backends = append(backends, pagedBackends...)
	}

	if err := pageIter.Error(); err != nil {
		return diag.FromErr(err)
	}

	// Filter backends
	var filteredBackends []objectstoragev2.S3Backend
	for _, backend := range backends {
		// Filter by enabled status if requested
		if enabledOnly && (backend.Enabled == nil || !*backend.Enabled) {
			continue
		}

		// Filter by name if provided
		if nameFilter != "" && !contains(backend.Name, nameFilter) {
			continue
		}

		filteredBackends = append(filteredBackends, backend)
	}

	// Convert to terraform data structure
	backendList := make([]interface{}, len(filteredBackends))
	for i, backend := range filteredBackends {
		backendMap := map[string]interface{}{
			"identifier":   backend.Identifier,
			"name":         backend.Name,
			"endpoint":     backend.Endpoint.Identifier,
			"backend_user": backend.BackendUser,
		}

		// Handle optional fields
		if backend.BackendType != nil {
			backendMap["backend_type"] = backend.BackendType.Identifier
		}
		if backend.Enabled != nil {
			backendMap["enabled"] = *backend.Enabled
		}
		if backend.State != nil {
			backendMap["state"] = backend.State.String()
		}

		// Handle organization references
		if backend.Reseller != "" {
			backendMap["reseller"] = backend.Reseller
		}
		if backend.Customer != "" {
			backendMap["customer"] = backend.Customer
		}

		// Note: CreatedAt and UpdatedAt fields are not available in the current API structure

		backendList[i] = backendMap
	}

	if err := d.Set("backends", backendList); err != nil {
		return diag.FromErr(err)
	}

	// Set ID based on time to make resource unique
	d.SetId(generateDataSourceID())

	return diags
}

// Helper function to check if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(substr) == 0 || len(s) >= len(substr) &&
		(s == substr || stringContainsIgnoreCase(s, substr))
}

func stringContainsIgnoreCase(s, substr string) bool {
	// Simple case-insensitive contains check
	sLower := strings.ToLower(s)
	substrLower := strings.ToLower(substr)
	return strings.Contains(sLower, substrLower)
}
