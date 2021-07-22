package anxcloud

import (
	"context"
	"fmt"
	"time"

	"github.com/anexia-it/go-anxcloud/pkg/client"
	"github.com/anexia-it/go-anxcloud/pkg/ipam/address"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	ipAddressStatusInactive = "Inactive"
	ipAddressStatusActive   = "Active"
	ipAddressStatusDeleted  = "Marked for deletion"
)

func resourceIPAddress() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIPAddressCreate,
		ReadContext:   resourceIPAddressRead,
		UpdateContext: resourceIPAddressUpdate,
		DeleteContext: resourceIPAddressDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Read:   schema.DefaultTimeout(1 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: schemaIPAddress(),
	}
}

func resourceIPAddressCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.Client)
	a := address.NewAPI(c)
	prefixID := d.Get("network_prefix_id").(string)

	def := address.Create{
		PrefixID:            prefixID,
		Address:             d.Get("address").(string),
		DescriptionCustomer: d.Get("description_customer").(string),
		Role:                d.Get("role").(string),
		Organization:        d.Get("organization").(string),
	}
	res, err := a.Create(ctx, def)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(res.ID)

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		info, err := a.Get(ctx, d.Id())
		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("unable to get ip address with '%s' id: %w", d.Id(), err))
		}
		if info.Status == ipAddressStatusInactive || info.Status == ipAddressStatusActive {
			return nil
		}
		return resource.RetryableError(fmt.Errorf("waiting for ip address with '%s' id to be ready", d.Id()))
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceIPAddressRead(ctx, d, m)
}

func resourceIPAddressRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags []diag.Diagnostic

	c := m.(client.Client)
	a := address.NewAPI(c)

	info, err := a.Get(ctx, d.Id())
	if err != nil {
		if err := handleNotFoundError(err); err != nil {
			return diag.FromErr(err)
		}
		d.SetId("")
		return nil
	}

	if err := d.Set("address", info.Name); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("network_prefix_id", info.PrefixID); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("status", info.Status); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("description_customer", info.DescriptionCustomer); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("description_internal", info.DescriptionInternal); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	// TODO: API require 'role' arg and returns 'role_text' arg, this must be fixed
	if err := d.Set("role", info.Role); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("version", info.Version); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("vlan_id", info.VLANID); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	return diags
}

func resourceIPAddressUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.Client)
	a := address.NewAPI(c)

	if !d.HasChanges("description_customer", "role") {
		return nil
	}

	def := address.Update{
		DescriptionCustomer: d.Get("description_customer").(string),
		Role:                d.Get("role").(string),
	}
	if _, err := a.Update(ctx, d.Id(), def); err != nil {
		return diag.FromErr(err)
	}

	return resourceIPAddressRead(ctx, d, m)
}

func resourceIPAddressDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.Client)
	a := address.NewAPI(c)

	err := a.Delete(ctx, d.Id())
	if err != nil {
		if err := handleNotFoundError(err); err != nil {
			return diag.FromErr(err)
		}
		d.SetId("")
		return nil
	}

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		info, err := a.Get(ctx, d.Id())
		if err != nil {
			if err := handleNotFoundError(err); err != nil {
				return resource.NonRetryableError(fmt.Errorf("unable to get ip address with id '%s': %w", d.Id(), err))
			}
			d.SetId("")
			return nil
		}
		if info.Status == ipAddressStatusDeleted {
			d.SetId("")
			return nil
		}
		return resource.RetryableError(fmt.Errorf("waiting for ip address with id '%s' to be deleted", d.Id()))
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
