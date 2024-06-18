package anxcloud

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"go.anx.io/go-anxcloud/pkg/vlan"
)

func dataSourceVLAN() *schema.Resource {
	return &schema.Resource{
		Description: "Provides details about an Anexia Cloud VLAN. This data source is useful if you want to use a non-terraform managed VLAN.",
		ReadContext: dataSourceVLANRead,
		Schema: map[string]*schema.Schema{
			"id":   {Type: schema.TypeString, Optional: true, Description: "The VLAN ID.", ExactlyOneOf: []string{"id", "name"}},
			"name": {Type: schema.TypeString, Optional: true, Description: "VLAN name."},

			"location_id":          {Type: schema.TypeString, Computed: true, Description: "ANX Location Identifier."},
			"vm_provisioning":      {Type: schema.TypeBool, Computed: true, Description: "True if VM provisioning shall be enabled. Defaults to false."},
			"description_customer": {Type: schema.TypeString, Computed: true, Description: "Additional customer description."},
			"description_internal": {Type: schema.TypeString, Computed: true, Description: "Internal description."},
			"role_text":            {Type: schema.TypeString, Computed: true, Description: "Role of the VLAN."},
			"status":               {Type: schema.TypeString, Computed: true, Description: "VLAN status."},
			"locations":            schemaLocations(),
		},
	}
}

func dataSourceVLANRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(providerContext).legacyClient
	v := vlan.NewAPI(c)

	// The schema ensures that either ID or name have to be set. We first check whether the identifier
	// has been set and if so, return the VLAN for that.
	// Otherwise, we fall back to search by name.

	if idFromScheme, ok := d.GetOk("id"); ok {
		d.SetId(idFromScheme.(string))

		if diags := resourceVLANRead(ctx, d, m); diags.HasError() {
			return diags
		}

		// Since the resourceVLANRead implementation handles the case of "not found" by emptying out the ID,
		// we check whether that's the case and handle it as an error.
		if d.Id() == "" {
			return diag.Errorf(
				`No VLAN with the given identifier %q could be found.
If you are sure that it exists, verify that you have the correct permissions to access it.`, idFromScheme)
		}
		return nil
	}

	// Since we didn't got an ID, we search for the name instead.
	var foundVLANID string
	name := d.Get("name").(string)

	vlans, err := listAllPages(func(page int) ([]vlan.Summary, error) {
		return v.List(ctx, page, 100, url.QueryEscape(name))
	})
	if err != nil {
		return diag.FromErr(fmt.Errorf("querying VLAN with name %q from engine: %w", name, err))
	}

	// Because searches like "VLAN123" would also match "VLAN1234", we iterate over the results and search for
	// an exact match.
	for _, v := range vlans {
		if v.Name == name {
			// Although VLAN identifiers should be unique, we should not rely on that. In case there's an ambiguity,
			// better report an error instead of accidentally breaking infrastructure.
			if foundVLANID != "" {
				return diag.Errorf("Name ambiguity detected when searching for VLAN with name %q. You should reference the VLAN using one of its identifiers (%s) instead of relying on the name.",
					name,
					strings.Join([]string{foundVLANID, v.Identifier}, ", "))
			}

			foundVLANID = v.Identifier
		}
	}

	// If the ID is still empty, that means that we haven't found a suitable VLAN.
	if foundVLANID == "" {
		return diag.Errorf("No VLAN found with the name %q.", name)
	}

	// Once we determined our identifier, we set it and reuse our code for fetching the detailed information.
	d.SetId(foundVLANID)
	return resourceVLANRead(ctx, d, m)
}
