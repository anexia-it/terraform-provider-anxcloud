package anxcloud

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"go.anx.io/go-anxcloud/pkg/core/location"
)

func dataSourceCoreLocations() *schema.Resource {
	return &schema.Resource{
		Description: "Provides available locations.",
		ReadContext: dataSourceCoreLocationsRead,
		Schema:      schemaDataSourceLocations(),
	}
}

func dataSourceCoreLocationsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(providerContext).legacyClient
	l := location.NewAPI(c)

	locations, err := listAllPages(func(page int) ([]location.Location, error) {
		return l.List(ctx, page, 100, d.Get("search").(string))
	})
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("locations", flattenCoreLocations(locations)); err != nil {
		return diag.FromErr(err)
	}

	if s := d.Get("search").(string); s != "" {
		d.SetId(strconv.FormatInt(time.Now().Round(time.Hour).Unix(), 10) + "-" + s)
	} else {
		d.SetId(strconv.FormatInt(time.Now().Round(time.Hour).Unix(), 10))
	}

	return nil
}
