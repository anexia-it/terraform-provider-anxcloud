package anxcloud

import (
	"context"
	"net"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"go.anx.io/go-anxcloud/pkg/ipam/address"
	"go.anx.io/go-anxcloud/pkg/utils/param"
)

func dataSourceIPAddress() *schema.Resource {
	return &schema.Resource{
		Description: `
Retrieves an IP address.

### Known limitations

- When using the address argument, only IP addresses unique to the scope of your access token for Anexia Cloud can be retrieved. You can however get a unique result by specifying the related VLAN or network prefix.
`,
		ReadContext: dataSourceIPAddressRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Description:  identifierDescription,
				ExactlyOneOf: []string{"id", "address"},
			},
			"network_prefix_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Identifier of the related network prefix.",
			},
			"address": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "IP address.",
			},
			"vlan_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The associated VLAN identifier.",
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
			"role": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Role of the IP address",
			},
			"organization": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Customer of yours. Reseller only.",
			},
			"version": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "IP version.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status of the IP address",
			},
		},
	}
}

func dataSourceIPAddressRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(providerContext).legacyClient
	a := address.NewAPI(c)

	var id string

	if addressFromSchema, ok := d.GetOk("address"); ok {
		// Parse IP address to normalize it (engine uses shortened, lowercase IPv6 address names)
		parsedAddr := net.ParseIP(addressFromSchema.(string))
		if parsedAddr == nil {
			return diag.Errorf("Failed to parse IP address.")
		}

		parsedAddrString := parsedAddr.String()

		filters := []param.Parameter{
			param.ParameterBuilder("search")(parsedAddrString),
		}

		if vlanID, ok := d.GetOk("vlan_id"); ok {
			filters = append(filters, address.VlanFilter(vlanID.(string)))
		}

		if networkPrefixID, ok := d.GetOk("network_prefix_id"); ok {
			filters = append(filters, address.PrefixFilter(networkPrefixID.(string)))
		}

		res, err := listAllPages(func(page int) ([]address.Summary, error) {
			return a.GetFiltered(ctx, page, 100, filters...)
		})
		if err != nil {
			return diag.FromErr(err)
		}

		for _, entry := range res {
			// we need an exact match, because search of 1.2.3.4 will also yield XX1.2.3.4X in results
			if entry.Name == parsedAddrString {
				if id != "" {
					// this might happen with private IPs in separate VLANs
					return diag.Errorf("IP address was found multiple times.")
				}
				id = entry.ID
			}
		}

		if id == "" {
			return diag.Errorf("IP address was not found.")
		}
	} else {
		id = d.Get("id").(string)
	}

	addr, err := a.Get(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(addr.ID)

	var diags []diag.Diagnostic

	if err = d.Set("network_prefix_id", addr.PrefixID); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err = d.Set("address", addr.Name); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err = d.Set("description_customer", addr.DescriptionCustomer); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err = d.Set("description_internal", addr.DescriptionInternal); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err = d.Set("role", addr.Role); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err = d.Set("version", addr.Version); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err = d.Set("vlan_id", addr.VLANID); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err = d.Set("status", addr.Status); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	return diags
}
