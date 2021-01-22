package anxcloud

import (
	"context"

	"github.com/anexia-it/go-anxcloud/pkg/client"
	"github.com/anexia-it/go-anxcloud/pkg/vsphere/provisioning/nictype"

	"github.com/google/uuid"
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

	if id := uuid.New().String(); id != "" {
		d.SetId(id)
		return nil
	}

	return diag.Errorf("unable to create uuid for IPs data source")
}
