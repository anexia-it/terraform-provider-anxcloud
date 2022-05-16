package anxcloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"go.anx.io/go-anxcloud/pkg/vsphere/provisioning/templates"
)

func dataSourceTemplate() *schema.Resource {
	return &schema.Resource{
		Description: "Provides available templates for specified location. This information can be used to provision virtual servers using the `anxcloud_virtual_server` resource.",
		ReadContext: dataSourceTemplateRead,
		Schema:      schemaTemplate(),
	}
}

func dataSourceTemplateRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(providerContext).legacyClient
	t := templates.NewAPI(c)
	locationID := d.Get("location_id").(string)
	templateType := d.Get("template_type").(string)
	templates, err := t.List(ctx, locationID, templateType, 1, 1000)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("templates", flattenTemplates(templates)); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(locationID)
	return nil
}
