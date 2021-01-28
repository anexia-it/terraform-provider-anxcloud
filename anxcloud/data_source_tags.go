package anxcloud

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/anexia-it/go-anxcloud/pkg/client"
	"github.com/anexia-it/go-anxcloud/pkg/core/tags"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTags() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceTagsRead,
		Schema:      schemaTags(),
	}
}

func dataSourceTagsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.Client)
	tagsAPI := tags.NewAPI(c)

	page := d.Get("page").(int)
	limit := d.Get("limit").(int)
	query := d.Get("query").(string)
	serviceIdentifier := d.Get("service_identifier").(string)
	organizationIdentifier := d.Get("organization_identifier").(string)
	order := d.Get("order").(string)
	sortAscending := d.Get("sort_ascending").(bool)

	tags, err := tagsAPI.List(ctx, page, limit, query, serviceIdentifier, organizationIdentifier, order, sortAscending)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("tags", flattenTags(tags)); err != nil {
		return diag.FromErr(err)
	}

	id := fmt.Sprintf("%s-%s-%s-%s",
		strconv.FormatInt(time.Now().Round(time.Hour).Unix(), 10), query, serviceIdentifier, organizationIdentifier)
	d.SetId(id)
	return nil
}
