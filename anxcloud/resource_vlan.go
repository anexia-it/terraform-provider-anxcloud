package anxcloud

import (
	"context"
	"fmt"
	"time"

	"github.com/anexia-it/go-anxcloud/pkg/client"
	"github.com/anexia-it/go-anxcloud/pkg/vlan"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	vlanStatusActive = "Active"
)

func resourceVLAN() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVLANCreate,
		ReadContext:   resourceVLANRead,
		UpdateContext: resourceVLANUpdate,
		DeleteContext: resourceVLANDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Read:   schema.DefaultTimeout(1 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: schemaVLAN(),
	}
}

func resourceVLANCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.Client)
	v := vlan.NewAPI(c)
	locationID := d.Get("location_id").(string)

	def := vlan.CreateDefinition{
		Location:            locationID,
		VMProvisioning:      d.Get("vm_provisioning").(bool),
		CustomerDescription: d.Get("description_customer").(string),
	}
	res, err := v.Create(ctx, def)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(res.Identifier)

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		vlan, err := v.Get(ctx, d.Id())
		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("unable to fetch vlan with '%s' id: %w", d.Id(), err))
		}
		if vlan.Status == vlanStatusActive {
			return nil
		}
		return resource.RetryableError(fmt.Errorf("waiting for vlan with '%s' id to be '%s'", d.Id(), vlanStatusActive))
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceVLANRead(ctx, d, m)
}

func resourceVLANRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags []diag.Diagnostic
	c := m.(client.Client)
	v := vlan.NewAPI(c)

	vlan, err := v.Get(ctx, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("name", vlan.Name); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("role_text", vlan.Role); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("status", vlan.Status); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("description_customer", vlan.CustomerDescription); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("description_internal", vlan.InternalDescription); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("locations", flattenVLANLocations(vlan.Locations)); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	return diags
}

func resourceVLANUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.Client)
	v := vlan.NewAPI(c)

	if !d.HasChange("description_customer") {
		return nil
	}

	def := vlan.UpdateDefinition{
		CustomerDescription: d.Get("description_customer").(string),
	}
	if err := v.Update(ctx, d.Id(), def); err != nil {
		return diag.FromErr(err)
	}

	return resourceVLANRead(ctx, d, m)
}

func resourceVLANDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.Client)
	v := vlan.NewAPI(c)

	err := v.Delete(ctx, d.Id())
	if err != nil {
		if err := handleNotFoundError(err); err != nil {
			return diag.FromErr(err)
		}
		d.SetId("")
		return nil
	}

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err := v.Get(ctx, d.Id())
		if err != nil {
			if err := handleNotFoundError(err); err != nil {
				return resource.NonRetryableError(fmt.Errorf("unable to get vlan with id '%s': %w", d.Id(), err))
			}
			d.SetId("")
			return nil
		}
		return resource.RetryableError(fmt.Errorf("waiting for vlan with id '%s' to be deleted", d.Id()))
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
