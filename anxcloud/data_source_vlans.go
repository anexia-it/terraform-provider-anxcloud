package anxcloud

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"go.anx.io/go-anxcloud/pkg/vlan"
)

func dataSourceVLANs() *schema.Resource {
	return &schema.Resource{
		Description: "Provides available VLANs.",
		ReadContext: dataSourceVLANsRead,
		Schema:      schemaVLANs(),
	}
}

func dataSourceVLANsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(providerContext).legacyClient
	v := vlan.NewAPI(c)

	vlans, err := v.List(ctx, d.Get("page").(int), d.Get("limit").(int), d.Get("search").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("vlans", flattenVLANs(vlans)); err != nil {
		return diag.FromErr(err)
	}

	if s := d.Get("search").(string); s != "" {
		d.SetId(strconv.FormatInt(time.Now().Round(time.Hour).Unix(), 10) + "-" + s)
	} else {
		d.SetId(strconv.FormatInt(time.Now().Round(time.Hour).Unix(), 10))
	}

	return nil
}
