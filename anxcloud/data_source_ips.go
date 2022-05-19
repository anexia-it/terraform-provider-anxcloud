package anxcloud

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"go.anx.io/go-anxcloud/pkg/ipam/address"
)

func dataSourceIPAddresses() *schema.Resource {
	return &schema.Resource{
		Description: "Provides available IP addresses.",
		ReadContext: dataSourceIPAddressesRead,
		Schema:      schemaIPAddresses(),
	}
}

func dataSourceIPAddressesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(providerContext).legacyClient
	a := address.NewAPI(c)

	addresses, err := listAllPages(func(page int) ([]address.Summary, error) {
		return a.List(ctx, page, 100, d.Get("search").(string))
	})
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
