package anxcloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"go.anx.io/go-anxcloud/pkg/api"
	"go.anx.io/go-anxcloud/pkg/api/types"
	clouddnsv1 "go.anx.io/go-anxcloud/pkg/apis/clouddns/v1"
)

func dataSourceDNSRecords() *schema.Resource {
	return &schema.Resource{
		Description: "Provides DNS records for a specified zone.",
		ReadContext: dataSourceDNSRecordsRead,
		Schema:      schemaDNSRecords(),
	}
}

func dataSourceDNSRecordsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	a := apiFromProviderConfig(m)

	zoneName := d.Get("zone_name").(string)

	var pageIter types.PageInfo
	recordList := clouddnsv1.Record{ZoneName: zoneName}
	err := a.List(ctx, &recordList, api.Paged(1, 100, &pageIter))
	if err != nil {
		return diag.FromErr(err)
	}

	records := make([]clouddnsv1.Record, 0, pageIter.TotalItems())
	var pagedRecords []clouddnsv1.Record
	for pageIter.Next(&pagedRecords) {
		records = append(records, pagedRecords...)
	}

	if err := pageIter.Error(); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("records", flattenDNSRecordsV1(records)); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(zoneName)
	return nil
}
