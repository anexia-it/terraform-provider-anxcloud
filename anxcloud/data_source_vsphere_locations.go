package anxcloud

import (
	"context"

	"github.com/anexia-it/go-anxcloud/pkg/client"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/anexia-it/go-anxcloud/pkg/vsphere/provisioning/location"
)

func dataSourceVSphereLocations() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVSphereLocationsRead,
		Schema:      schemaDataSourceVSPhereLocations(),
	}
}

func dataSourceVSphereLocationsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.Client)
	api := location.NewAPI(c)

	page := d.Get("page").(int)
	limit := d.Get("limit").(int)
	locationCode := d.Get("location_code").(string)
	organization := d.Get("organization").(string)
	locations, err := api.List(ctx, page, limit, locationCode, organization)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("locations", flattenVSphereLocations(locations)); err != nil {
		return diag.FromErr(err)
	}

	if id := uuid.New().String(); id != "" {
		d.SetId(id)
		return nil
	}

	return diag.Errorf("unable to create uuid for locations data source")
}
