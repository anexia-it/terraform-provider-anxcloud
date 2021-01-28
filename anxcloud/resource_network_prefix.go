package anxcloud

import (
	"context"
	"fmt"
	"time"

	"github.com/anexia-it/go-anxcloud/pkg/client"
	"github.com/anexia-it/go-anxcloud/pkg/ipam/prefix"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	prefixStatusActive  = "Active"
	prefixStatusDeleted = "Marked for deletion"
)

func resourceNetworkPrefix() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNetworkPrefixCreate,
		ReadContext:   resourceNetworkPrefixRead,
		UpdateContext: resourceNetworkPrefixUpdate,
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
		Schema: schemaNetworkPrefix(),
	}
}

func resourceNetworkPrefixCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.Client)
	p := prefix.NewAPI(c)
	locationID := d.Get("location_id").(string)

	createParams := prefix.Create{
		Location:                locationID,
		IPVersion:               d.Get("ip_version").(int),
		Type:                    d.Get("type").(int),
		NetworkMask:             d.Get("netmask").(int),
		CreateVLAN:              d.Get("new_vlan").(bool),
		VLANID:                  d.Get("vlan_id").(string),
		EnableRedundancy:        d.Get("router_redundancy").(bool),
		EnableVMProvisioning:    d.Get("vm_provisioning").(bool),
		CustomerDescription:     d.Get("description_customer").(string),
		CustomerVLANDescription: d.Get("description_vlan_customer").(string),
		Organization:            d.Get("organization").(string),
	}
	res, err := p.Create(ctx, createParams)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(res.ID)

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		pref, err := p.Get(ctx, d.Id())
		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("unable to get network prefix with '%s' id", d.Id()))
		}
		if pref.Status == prefixStatusActive {
			return nil
		}
		return resource.RetryableError(fmt.Errorf("waiting for network prefix with '%s' id to be: %s", d.Id(), prefixStatusActive))
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceNetworkPrefixRead(ctx, d, m)
}

func resourceNetworkPrefixRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags []diag.Diagnostic

	c := m.(client.Client)
	p := prefix.NewAPI(c)

	info, err := p.Get(ctx, d.Id())
	if err != nil {
		if err := handleNotFoundError(err); err != nil {
			return diag.FromErr(err)
		}
		d.SetId("")
		return nil
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
	if err := d.Set("locations", flattenNetworkPrefixLocations(info.Locations)); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	return diags
}

func resourceNetworkPrefixUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.Client)
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
	c := m.(client.Client)
	p := prefix.NewAPI(c)

	if err := p.Delete(ctx, d.Id()); err != nil {
		if err := handleNotFoundError(err); err != nil {
			return diag.FromErr(err)
		}
		d.SetId("")
		return nil
	}

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		info, err := p.Get(ctx, d.Id())
		if err != nil {
			if err := handleNotFoundError(err); err != nil {
				return resource.NonRetryableError(fmt.Errorf("unable to get network prefix with id '%s': %w", d.Id(), err))
			}
			d.SetId("")
			return nil
		}
		if info.Status == prefixStatusDeleted {
			d.SetId("")
			return nil
		}
		return resource.RetryableError(fmt.Errorf("waiting for network prefix with id '%s' to be deleted", d.Id()))
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
