package anxcloud

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"go.anx.io/go-anxcloud/pkg/core/tags"
)

func dataSourceTags() *schema.Resource {
	return &schema.Resource{
		Description: "Provides available service tags.",
		ReadContext: dataSourceTagsRead,
		Schema:      schemaTags(),
	}
}

func dataSourceTagsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(providerContext).legacyClient
	tagsAPI := tags.NewAPI(c)

	query := d.Get("query").(string)
	serviceIdentifier := d.Get("service_identifier").(string)
	organizationIdentifier := d.Get("organization_identifier").(string)
	order := d.Get("order").(string)
	sortAscending := d.Get("sort_ascending").(bool)

	tags, err := listAllPages(func(page int) ([]tags.Summary, error) {
		return tagsAPI.List(ctx, page, 100, query, serviceIdentifier, organizationIdentifier, order, sortAscending)
	})
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("tags", flattenTags(tags)); err != nil {
		return diag.FromErr(err)
	}

	id := strconv.FormatInt(time.Now().Round(time.Hour).Unix(), 10)
	if len(query) > 0 {
		id = fmt.Sprintf("%s-%s", id, query)
	}
	if len(serviceIdentifier) > 0 {
		id = fmt.Sprintf("%s-%s", id, serviceIdentifier)
	}
	if len(organizationIdentifier) > 0 {
		id = fmt.Sprintf("%s-%s", id, organizationIdentifier)
	}
	d.SetId(id)
	return nil
}
