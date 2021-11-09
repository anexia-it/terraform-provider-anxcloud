package anxcloud

import (
	"context"
	"github.com/anexia-it/go-anxcloud/pkg/client"
	"github.com/anexia-it/go-anxcloud/pkg/clouddns/zone"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDNSRecords() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDNSRecordsRead,
		Schema:      schemaDNSRecords(),
	}
}

func dataSourceDNSRecordsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.Client)
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
