package anxcloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"go.anx.io/go-anxcloud/pkg/clouddns/zone"
)

func dataSourceDNSRecords() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDNSRecordsRead,
		Schema:      schemaDNSRecords(),
	}
}

func dataSourceDNSRecordsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(providerContext).legacyClient
	a := zone.NewAPI(c)

	zoneName := d.Get("zone_name").(string)

	records, err := a.ListRecords(ctx, zoneName)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("records", flattenDnsRecords(records)); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(zoneName)
	return nil
}
