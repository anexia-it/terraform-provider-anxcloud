package anxcloud

import (
	"context"

	"github.com/anexia-it/go-anxcloud/pkg/client"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	cpuperformancetypes "github.com/anexia-it/go-anxcloud/pkg/vsphere/provisioning/cpuperformancetypes"
)

func dataSourceCPUPerformanceTypes() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceCPUPerformanceTypesRead,
		Schema:      schemaCPUPerformanceType(),
	}
}

func dataSourceCPUPerformanceTypesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.Client)
	api := cpuperformancetypes.NewAPI(c)

	types, err := api.List(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("types", flattenCPUPerformanceTypes(types)); err != nil {
		return diag.FromErr(err)
	}

	if id := uuid.New().String(); id != "" {
		d.SetId(id)
		return nil
	}

	return diag.Errorf("unable to create uuid for cpu performance types data source")
}
