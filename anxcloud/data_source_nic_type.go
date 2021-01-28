package anxcloud

import (
	"context"
	"strconv"
	"time"

	"github.com/anexia-it/go-anxcloud/pkg/client"
	"github.com/anexia-it/go-anxcloud/pkg/vsphere/provisioning/nictype"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNICTypes() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceNICTypesRead,
		Schema:      schemaNICTypes(),
	}
}

func dataSourceNICTypesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.Client)
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
