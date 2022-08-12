package anxcloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	corev1 "go.anx.io/go-anxcloud/pkg/apis/core/v1"
)

func dataSourceCoreLocation() *schema.Resource {
	return &schema.Resource{
		Description: "Retrieves a location identified by it's `identifier` or human-readable `code` as selectable in the Engine. " +
			"This data source can be used to lookup a locations `identifier` required by other resources and data sources available in this provider.",
		ReadContext: dataSourceCoreLocationRead,
		Schema: schemaWith(schemaLocation(),
			fieldsExactlyOneOf("identifier", "code"),
		),
	}
}

func dataSourceCoreLocationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	a := apiFromProviderConfig(m)

	location := corev1.Location{
		Identifier: d.Get("identifier").(string),
		Code:       d.Get("code").(string),
	}

	if err := a.Get(ctx, &location); err != nil {
		return diag.FromErr(err)
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
