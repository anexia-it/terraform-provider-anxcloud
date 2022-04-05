package anxcloud

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"go.anx.io/go-anxcloud/pkg/vsphere/provisioning/nictype"
)

func dataSourceNICTypes() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceNICTypesRead,
		Schema:      schemaNICTypes(),
	}
}

func dataSourceNICTypesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(providerContext).legacyClient
	n := nictype.NewAPI(c)

	nicTypes, err := n.List(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("nic_types", nicTypes); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.FormatInt(time.Now().Round(time.Hour).Unix(), 10))
	return nil
}
