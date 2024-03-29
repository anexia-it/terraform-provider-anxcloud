package anxcloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"go.anx.io/go-anxcloud/pkg/ipam/prefix"
)

const (
	prefixStatusActive  = "Active"
	prefixStatusFailure = "Failed"
	prefixStatusDeleted = "Marked for deletion"
)

func resourceNetworkPrefix() *schema.Resource {
	return &schema.Resource{
		Description:   "This resource allows you to create and configure network prefix.",
		CreateContext: tagsMiddlewareCreate(resourceNetworkPrefixCreate),
		ReadContext:   tagsMiddlewareRead(resourceNetworkPrefixRead),
		UpdateContext: tagsMiddlewareUpdate(resourceNetworkPrefixUpdate),
		DeleteContext: resourceNetworkPrefixDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Read:   schema.DefaultTimeout(1 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: withTagsAttribute(schemaNetworkPrefix()),
	}
}

func resourceNetworkPrefixCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(providerContext).legacyClient
	p := prefix.NewAPI(c)
	locationID := d.Get("location_id").(string)

	createParams := prefix.Create{
		Location:            locationID,
		IPVersion:           d.Get("ip_version").(int),
		Type:                d.Get("type").(int),
		NetworkMask:         d.Get("netmask").(int),
		VLANID:              d.Get("vlan_id").(string),
		EnableRedundancy:    d.Get("router_redundancy").(bool),
		CustomerDescription: d.Get("description_customer").(string),
		Organization:        d.Get("organization").(string),
		CreateEmpty:         d.Get("create_empty").(bool),
	}
	res, err := p.Create(ctx, createParams)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(res.ID)

	err = retry.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *retry.RetryError {
		pref, err := p.Get(ctx, d.Id())
		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("unable to get network prefix with '%s' id", d.Id()))
		}
		if pref.Status == prefixStatusActive || pref.Status == prefixStatusFailure {
			return nil
		}
		return retry.RetryableError(fmt.Errorf("waiting for network prefix with '%s' id to be: %s", d.Id(), prefixStatusActive))
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceNetworkPrefixRead(ctx, d, m)
}

func resourceNetworkPrefixRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags []diag.Diagnostic

	c := m.(providerContext).legacyClient
	p := prefix.NewAPI(c)

	info, err := p.Get(ctx, d.Id())
	if err != nil {
		if err := handleNotFoundError(err); err != nil {
			return diag.FromErr(err)
		}
		d.SetId("")
		return nil
	}

	createEmpty := d.Get("create_empty")
	if err := d.Set("create_empty", createEmpty); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	// CIDR value is set in the 'name' field, this should be changed
	if err := d.Set("cidr", info.Name); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("ip_version", info.IPVersion); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("netmask", info.NetworkMask); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("description_customer", info.CustomerDescription); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("description_internal", info.InternalDescription); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("role_text", info.Role); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("status", info.Status); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("type", info.PrefixType); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("locations", flattenNetworkPrefixLocations(info.Locations)); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if len(info.Locations) > 0 {
		if err := d.Set("location_id", info.Locations[0].ID); err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
	}
	if err := d.Set("router_redundancy", info.RouterRedundancy); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	if len(info.Vlans) == 0 {
		diags = append(diags, diag.Errorf("no VLAN seems to be attached to prefix '%s'", info.ID)...)
	} else if err := d.Set("vlan_id", info.Vlans[0].ID); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	return diags
}

func resourceNetworkPrefixUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(providerContext).legacyClient
	p := prefix.NewAPI(c)

	if !d.HasChange("description_customer") {
		return nil
	}

	def := prefix.Update{
		CustomerDescription: d.Get("description_customer").(string),
	}
	if _, err := p.Update(ctx, d.Id(), def); err != nil {
		return diag.FromErr(err)
	}

	return resourceNetworkPrefixRead(ctx, d, m)
}

func resourceNetworkPrefixDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(providerContext).legacyClient
	p := prefix.NewAPI(c)

	if err := p.Delete(ctx, d.Id()); err != nil {
		if err := handleNotFoundError(err); err != nil {
			return diag.FromErr(err)
		}
		d.SetId("")
		return nil
	}

	err := retry.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *retry.RetryError {
		info, err := p.Get(ctx, d.Id())
		if err != nil {
			if err := handleNotFoundError(err); err != nil {
				return retry.NonRetryableError(fmt.Errorf("unable to get network prefix with id '%s': %w", d.Id(), err))
			}
			d.SetId("")
			return nil
		}
		if info.Status == prefixStatusDeleted {
			d.SetId("")
			return nil
		}
		return retry.RetryableError(fmt.Errorf("waiting for network prefix with id '%s' to be deleted", d.Id()))
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
