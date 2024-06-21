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

	var (
		id, _   = d.Get("id").(string)
		name, _ = d.Get("name").(string)
	)

	// The schema ensures that either ID or name have to be set. We first check
	// whether the identifier has been set and if so, return the VLAN for that.
	// Otherwise, we fall back to search by name.
	var (
		resultID string
		diags    diag.Diagnostics
	)

	switch {
	case id != "":
		resultID, diags = findVLANByID(ctx, v, id)
	case name != "":
		resultID, diags = findVLANByName(ctx, v, name)
	default:
		return diag.Errorf("Either provide a non-empty %q or %q to query a VLAN.", "id", "name")
	}

	if diags.HasError() {
		return diags
	} else if resultID == "" {
		return diag.Errorf(`An unexpected error occurred and the provider could not determine the identifier.
Please report this issue to the developers by opening a ticket on GitHub:
https://github.com/anexia-it/terraform-provider-anxcloud/issues/new`)
	}

	// Once we determined our identifier, we set it and reuse our code for fetching the detailed information.
	d.SetId(resultID)
	return resourceVLANRead(ctx, d, m)
}

// findVLANByName searches for a given VLAN by the provided name.
func findVLANByName(ctx context.Context, v vlan.API, name string) (string, diag.Diagnostics) {
	var foundID string

	vlans, err := listAllPages(func(page int) ([]vlan.Summary, error) {
		return v.List(ctx, page, 100, url.QueryEscape(name))
	})
	if err != nil {
		return "", diag.FromErr(fmt.Errorf("querying VLAN with name %q from engine: %w", name, err))
	}

	// Because searches like "VLAN123" would also match "VLAN1234", we iterate over
	// the results and search for an exact match.
	for _, v := range vlans {
		if v.Name == name {

			// Although VLAN identifiers should be unique, we should not rely on that. In
			// case there's an ambiguity, better report an error instead of accidentally
			// breaking infrastructure.
			if foundID != "" {
				return "", diag.Errorf("Name ambiguity detected when searching for VLAN with name %q. You should reference the VLAN using one of its identifiers (%s) instead of relying on the name.",
					name,
					strings.Join([]string{foundID, v.Identifier}, ", "))
			}

			foundID = v.Identifier
		}
	}

	// If the ID is still empty, that means that we haven't found a suitable VLAN.
	var diags diag.Diagnostics
	if foundID == "" {
		diags = append(diags,
			diag.Errorf(`No VLAN found with the name %q.
If you are sure that it exists, verify that you have the correct permissions to access it.`, name)...)
	}

	return foundID, diags
}

// findVLANByID finds a VLAN by its ID and returns an error otherwise.
func findVLANByID(ctx context.Context, v vlan.API, id string) (string, diag.Diagnostics) {
	info, err := v.Get(ctx, id)
	if err != nil {
		if err := handleNotFoundError(err); err != nil {
			return "", diag.FromErr(err)
		}

		return "", diag.Errorf(
			`No VLAN with the given identifier %q could be found.
If you are sure that it exists, verify that you have the correct permissions to access it.`, id)
	}

	return info.Identifier, nil
}
