package anxcloud

import (
	"context"

	"github.com/anexia-it/go-anxcloud/pkg/client"
	"github.com/anexia-it/go-anxcloud/pkg/core/location"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceCoreLocations() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceCoreLocationsRead,
		Schema:      schemaDataSourceLocation(),
	}
}

func dataSourceCoreLocationsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.Client)
	n := location.NewAPI(c)

	locations, err := n.List(ctx, d.Get("page").(int), d.Get("limit").(int), d.Get("search").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("locations", flattenCoreLocations(locations)); err != nil {
		return diag.FromErr(err)
	}

	if id := uuid.New().String(); id != "" {
		d.SetId(id)
		return nil
	}

	return diag.Errorf("unable to create uuid for core locations data source")
}
