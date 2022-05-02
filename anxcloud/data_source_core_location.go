package anxcloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"go.anx.io/go-anxcloud/pkg/api"
	"go.anx.io/go-anxcloud/pkg/api/types"
	corev1 "go.anx.io/go-anxcloud/pkg/apis/core/v1"
)

func dataSourceCoreLocation() *schema.Resource {
	return &schema.Resource{
		Description: "Retrieves Location identified by it's `code` as selectable in the Engine.",
		ReadContext: dataSourceCoreLocationRead,
		Schema:      schemaLocation(),
	}
}

func dataSourceCoreLocationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	a := m.(providerContext).api

	code, exists := d.GetOk("code")
	if !exists {
		return diag.Errorf("location data-source requires code argument")
	}

	listCTX, cancel := context.WithCancel(ctx)
	defer cancel()
	searchLocation := &corev1.Location{Code: code.(string)}
	channel := make(types.ObjectChannel)
	if err := a.List(listCTX, searchLocation, api.ObjectChannel(&channel)); err != nil {
		return diag.FromErr(err)
	}

	found := false
	location := &corev1.Location{}
	for res := range channel {
		if err := res(location); err != nil {
			return diag.FromErr(err)
		}
		if location.Code == searchLocation.Code {
			found = true
			break
		}
	}

	if !found {
		return diag.Errorf("location with specified code not found")
	}

	var err error
	var diags []diag.Diagnostic

	d.SetId(location.Identifier)
	if err = d.Set("identifier", location.Identifier); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err = d.Set("name", location.Name); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err = d.Set("code", location.Code); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err = d.Set("country", location.CountryCode); err != nil {
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
