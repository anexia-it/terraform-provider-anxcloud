package anxcloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"go.anx.io/go-anxcloud/pkg/client"
	"go.anx.io/go-anxcloud/pkg/vsphere/provisioning/disktype"
)

func dataSourceDiskTypes() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDiskTypesRead,
		Schema:      schemaDiskTypes(),
	}
}

func dataSourceDiskTypesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.Client)
	t := disktype.NewAPI(c)
	locationID := d.Get("location_id").(string)
	diskTypes, err := t.List(ctx, locationID, 0, 1000)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("types", flattenDiskTypes(diskTypes)); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(locationID)
	return nil
}
