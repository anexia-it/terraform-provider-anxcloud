package anxcloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"go.anx.io/go-anxcloud/pkg/core/location"
)

func dataSourceCoreLocation() *schema.Resource {
	return &schema.Resource{
		Description: "Retrieves a location identified by it's `code` as selectable in the Engine. Use this data source to specify the location identifier on other resources and data sources available in this provider.",
		ReadContext: dataSourceCoreLocationRead,
		Schema:      schemaLocation(),
	}
}

func dataSourceCoreLocationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(providerContext).legacyClient
	l := location.NewAPI(c)

	code, exists := d.GetOk("code")
	if !exists {
		return diag.Errorf("location data-source requires code argument")
	}

	location, err := l.GetByCode(ctx, code.(string))
	if err != nil {
		return diag.FromErr(err)
	}

	var diags []diag.Diagnostic

	d.SetId(location.ID)
	if err = d.Set("identifier", location.ID); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err = d.Set("name", location.Name); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err = d.Set("code", location.Code); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err = d.Set("country", location.Country); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err = d.Set("city_code", location.CityCode); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err = d.Set("lat", location.Latitude); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err = d.Set("lon", location.Longitude); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	return diags
}
