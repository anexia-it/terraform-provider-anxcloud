package anxcloud

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/anexia-it/go-anxcloud/pkg/client"
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

	id := strconv.FormatInt(time.Now().Round(time.Hour).Unix(), 10)
	if len(locationCode) > 0 {
		id = fmt.Sprintf("%s-%s", id, locationCode)
	}
	if len(organization) > 0 {
		id = fmt.Sprintf("%s-%s", id, organization)
	}
	d.SetId(id)
	return nil
}
