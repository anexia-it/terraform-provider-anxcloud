package anxcloud

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"go.anx.io/go-anxcloud/pkg/clouddns/zone"
)

func datasourceDNSZones() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDNSZonesRead,
		Schema:      schemaDNSZones(),
	}
}

func dataSourceDNSZonesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(providerContext).legacyClient
	a := zone.NewAPI(c)

	zones, err := a.List(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("zones", flattenDnsZones(zones)); err != nil {
		return diag.FromErr(err)
	}

	id := strconv.FormatInt(time.Now().Round(time.Hour).Unix(), 10)
	d.SetId(id)
	return nil
}
