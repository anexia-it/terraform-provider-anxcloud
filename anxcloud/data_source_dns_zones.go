package anxcloud

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"go.anx.io/go-anxcloud/pkg/api"
	"go.anx.io/go-anxcloud/pkg/api/types"
	clouddnsv1 "go.anx.io/go-anxcloud/pkg/apis/clouddns/v1"
)

func datasourceDNSZones() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDNSZonesRead,
		Schema:      schemaDNSZones(),
	}
}

func dataSourceDNSZonesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	a := apiFromProviderConfig(m)

	var pageIter types.PageInfo
	err := a.List(ctx, &clouddnsv1.Zone{}, api.Paged(1, 100, &pageIter))
	if err != nil {
		return diag.FromErr(err)
	}

	zones := make([]clouddnsv1.Zone, 0, pageIter.TotalItems())
	var pagedZones []clouddnsv1.Zone
	for pageIter.Next(&pagedZones) {
		zones = append(zones, pagedZones...)
	}

	if err := pageIter.Error(); err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("zones", flattenDnsZones(zones)); err != nil {
		return diag.FromErr(err)
	}

	id := strconv.FormatInt(time.Now().Round(time.Hour).Unix(), 10)
	d.SetId(id)
	return nil
}
