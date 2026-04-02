package anxcloud

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"go.anx.io/go-anxcloud/pkg/vsphere/provisioning/availabilityzones"
)

func dataSourceAvailabilityZone() *schema.Resource {
	return &schema.Resource{
		Description: "Retrieves a Availability Zone identified by it's `name` for the specified location as selectable in the Engine. " +
			"This data source can be used to lookup a Availability Zone `identifier` required by other resources and data sources available in this provider.",
		ReadContext: dataSourceAvailabilityZoneRead,
		Schema: schemaWith(schemaAvailabilityZone(),
			fieldsExactlyOneOf("identifier", "name"),
			fieldsRequired("location_id"),
		),
	}
}

func schemaAvailabilityZone() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"identifier": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: identifierDescription,
		},
		"location_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Identifier of the location the Availability Zone is in.",
		},
		"name": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Availability Zone name.",
		},
	}
}

func dataSourceAvailabilityZoneRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(providerContext).legacyClient
	api := availabilityzones.NewAPI(c)

	identifier := d.Get("identifier").(string)
	locationID := d.Get("location_id").(string)
	name := d.Get("name").(string)

	//API call to get locations
	zones, err := api.List(ctx, locationID)
	if err != nil {
		return diag.FromErr(err)
	}

	//find exactly one zone where either identifier or name matches
	var zone availabilityzones.AvailabilityZone
	found := false
	for _, z := range zones {
		if z.Identifier == identifier || z.Name == name {

			if found {
				return diag.Errorf(
					"multiple Availability Zones found for name '%s'",
					name,
				)
			}

			found = true
			zone = z
		}
	}

	//none found -> abort
	if !found {
		var msg string
		if name != "" {
			msg = fmt.Sprintf("could not find Availability Zone with name '%s'", name)
		} else if identifier != "" {
			msg = fmt.Sprintf("could not find Availability Zone with ID '%s'", identifier)
		} else {
			msg = "Could not find Availability Zone"
		}
		err = errors.New(msg)
		return diag.FromErr(err)
	}

	var diags []diag.Diagnostic
	d.SetId(zone.Identifier)
	if err = d.Set("identifier", zone.Identifier); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err = d.Set("name", zone.Name); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	return diags
}
