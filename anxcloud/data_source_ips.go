package anxcloud

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"go.anx.io/go-anxcloud/pkg/client"
	"go.anx.io/go-anxcloud/pkg/ipam/address"
)

func dataSourceIPAddresses() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceIPAddressesRead,
		Schema:      schemaIPAddresses(),
	}
}

func dataSourceIPAddressesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.Client)
	a := address.NewAPI(c)

	addresses, err := a.List(ctx, d.Get("page").(int), d.Get("limit").(int), d.Get("search").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("addresses", flattenIPAddresses(addresses)); err != nil {
		return diag.FromErr(err)
	}

	if s := d.Get("search").(string); s != "" {
		d.SetId(strconv.FormatInt(time.Now().Round(time.Hour).Unix(), 10) + "-" + s)
	} else {
		d.SetId(strconv.FormatInt(time.Now().Round(time.Hour).Unix(), 10))
	}

	return nil
}
