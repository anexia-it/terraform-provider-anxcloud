package anxcloud

import (
	"context"

	"github.com/anexia-it/go-anxcloud/pkg/client"
	"github.com/anexia-it/go-anxcloud/pkg/vlan"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceVLANs() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVLANsRead,
		Schema:      schemaVLANs(),
	}
}

func dataSourceVLANsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.Client)
	v := vlan.NewAPI(c)

	vlans, err := v.List(ctx, d.Get("page").(int), d.Get("limit").(int), d.Get("search").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("vlans", flattenVLANs(vlans)); err != nil {
		return diag.FromErr(err)
	}

	if id := uuid.New().String(); id != "" {
		d.SetId(id)
		return nil
	}

	return diag.Errorf("unable to create uuid for vlan data source")
}
