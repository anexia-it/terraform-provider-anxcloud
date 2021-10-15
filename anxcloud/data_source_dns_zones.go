package anxcloud

import (
	"context"
	"github.com/anexia-it/go-anxcloud/pkg/client"
	"github.com/anexia-it/go-anxcloud/pkg/clouddns/zone"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceDNSZones() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDNSZonesRead,
		Schema: schemaDNSZones(),
	}
}

func dataSourceDNSZonesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.Client)
	a := zone.NewAPI(c)

	zones, err := a.List(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("zones", flattenDnsZones(zones)); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("?") // TODO what ID
	return nil
}