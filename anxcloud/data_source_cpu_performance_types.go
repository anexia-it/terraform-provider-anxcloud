package anxcloud

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	cpuperformancetypes "go.anx.io/go-anxcloud/pkg/vsphere/provisioning/cpuperformancetypes"
)

func dataSourceCPUPerformanceTypes() *schema.Resource {
	return &schema.Resource{
		Description: "Provides available cpu performance types. This information can be used to provision virtual servers using the `anxcloud_virtual_server` resource.",
		ReadContext: dataSourceCPUPerformanceTypesRead,
		Schema:      schemaCPUPerformanceType(),
	}
}

func dataSourceCPUPerformanceTypesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(providerContext).legacyClient
	api := cpuperformancetypes.NewAPI(c)

	types, err := api.List(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("types", flattenCPUPerformanceTypes(types)); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.FormatInt(time.Now().Round(time.Hour).Unix(), 10))
	return nil
}
