package anxcloud

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"go.anx.io/go-anxcloud/pkg/ipam/prefix"
)

func dataSourceNetworkPrefix() *schema.Resource {
	return &schema.Resource{
		Description: "Provides details about an Anexia Cloud network prefix. This data source is useful if you want to use a non-Terraform managed prefix.",
		ReadContext: dataSourceNetworkPrefixRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Network prefix identifier.",
				ExactlyOneOf: []string{"id", "cidr"},
			},
			"cidr": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Network prefix CIDR.",
				ExactlyOneOf: []string{"id", "cidr"},
			},
			"location_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Identifier of the prefix location.",
			},
			"netmask": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Network mask size.",
			},
			"ip_version": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "IP version: 4 for IPv4 or 6 for IPv6.",
			},
			"type": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Prefix type: 0 for public or 1 for private.",
			},
			"vlan_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Corresponding VLAN identifier.",
			},
			"router_redundancy": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether router redundancy is enabled.",
			},
			"description_customer": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Additional customer description.",
			},
			"description_internal": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Internal description.",
			},
			"role_text": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Role of the prefix.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Prefix status.",
			},
			"locations": schemaLocations(),
		},
	}
}

func dataSourceNetworkPrefixRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	prefixAPI := prefix.NewAPI(m.(providerContext).legacyClient)

	identifier := d.Get("id").(string)
	cidr := d.Get("cidr").(string)

	var diags diag.Diagnostics
	switch {
	case identifier != "":
		identifier, diags = findNetworkPrefixByID(ctx, prefixAPI, identifier)
	case cidr != "":
		identifier, diags = findNetworkPrefixByCIDR(ctx, prefixAPI, cidr)
	default:
		return diag.Errorf("Either provide a non-empty %q or %q to query a network prefix.", "id", "cidr")
	}
	if diags.HasError() {
		return diags
	}

	info, err := prefixAPI.Get(ctx, identifier)
	if err != nil {
		return diag.FromErr(fmt.Errorf("retrieving network prefix %q from engine: %w", identifier, err))
	}

	d.SetId(identifier)
	return setNetworkPrefixDataSourceData(d, info)
}

func findNetworkPrefixByID(ctx context.Context, prefixAPI prefix.API, identifier string) (string, diag.Diagnostics) {
	info, err := prefixAPI.Get(ctx, identifier)
	if err != nil {
		if err := handleNotFoundError(err); err != nil {
			return "", diag.FromErr(err)
		}

		return "", diag.Errorf(`No network prefix with the given identifier %q could be found.
If you are sure that it exists, verify that you have the correct permissions to access it.`, identifier)
	}

	return info.ID, nil
}

func findNetworkPrefixByCIDR(ctx context.Context, prefixAPI prefix.API, cidr string) (string, diag.Diagnostics) {
	prefixes, err := listAllPages(func(page int) ([]prefix.Summary, error) {
		return prefixAPI.List(ctx, page, 100, cidr)
	})
	if err != nil {
		return "", diag.FromErr(fmt.Errorf("querying network prefix with CIDR %q from engine: %w", cidr, err))
	}

	foundIDs := make([]string, 0, 1)
	for _, networkPrefix := range prefixes {
		if networkPrefix.Name == cidr {
			foundIDs = append(foundIDs, networkPrefix.ID)
		}
	}

	switch len(foundIDs) {
	case 0:
		return "", diag.Errorf(`No network prefix found with the CIDR %q.
If you are sure that it exists, verify that you have the correct permissions to access it.`, cidr)
	case 1:
		return foundIDs[0], nil
	default:
		return "", diag.Errorf("CIDR ambiguity detected when searching for network prefix %q. Reference the prefix using one of its identifiers (%s) instead.", cidr, strings.Join(foundIDs, ", "))
	}
}

func setNetworkPrefixDataSourceData(d *schema.ResourceData, info prefix.Info) diag.Diagnostics {
	var diags diag.Diagnostics
	set := func(name string, value interface{}) {
		if err := d.Set(name, value); err != nil {
			diags = append(diags, diag.FromErr(fmt.Errorf("setting network prefix field %q: %w", name, err))...)
		}
	}

	set("cidr", info.Name)
	set("ip_version", info.IPVersion)
	set("netmask", info.NetworkMask)
	set("description_customer", info.CustomerDescription)
	set("description_internal", info.InternalDescription)
	set("role_text", info.Role)
	set("status", info.Status)
	set("type", info.PrefixType)
	set("locations", flattenNetworkPrefixLocations(info.Locations))
	set("router_redundancy", info.RouterRedundancy)

	if len(info.Locations) > 0 {
		set("location_id", info.Locations[0].ID)
	}
	if len(info.Vlans) == 0 {
		diags = append(diags, diag.Errorf("no VLAN seems to be attached to prefix %q", info.ID)...)
	} else {
		set("vlan_id", info.Vlans[0].ID)
	}

	return diags
}
