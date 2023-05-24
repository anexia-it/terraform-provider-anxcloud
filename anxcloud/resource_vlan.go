package anxcloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"go.anx.io/go-anxcloud/pkg/vlan"
)

const (
	vlanStatusActive  = "Active"
	vlanStatusDeleted = "Marked for deletion"
)

func resourceVLAN() *schema.Resource {
	return &schema.Resource{
		Description:   "The VLAN resource allows you to create and configure VLAN.",
		CreateContext: tagsMiddlewareCreate(resourceVLANCreate),
		ReadContext:   tagsMiddlewareRead(resourceVLANRead),
		UpdateContext: tagsMiddlewareUpdate(resourceVLANUpdate),
		DeleteContext: resourceVLANDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Read:   schema.DefaultTimeout(1 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: withTagsAttribute(schemaVLAN()),
	}
}

func resourceVLANCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(providerContext).legacyClient
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

	err = retry.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *retry.RetryError {
		vlan, err := v.Get(ctx, d.Id())
		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("unable to fetch vlan with '%s' id: %w", d.Id(), err))
		}
		if vlan.Status == vlanStatusActive && vlan.VMProvisioning == def.VMProvisioning {
			return nil
		}
		return retry.RetryableError(fmt.Errorf("waiting for vlan with '%s' id to be '%s' and 'vm_provisioning' to have the desired state", d.Id(), vlanStatusActive))
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceVLANRead(ctx, d, m)
}

func resourceVLANRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags []diag.Diagnostic
	c := m.(providerContext).legacyClient
	v := vlan.NewAPI(c)

	vlan, err := v.Get(ctx, d.Id())
	if err != nil {
		if err := handleNotFoundError(err); err != nil {
			return diag.FromErr(err)
		}
		d.SetId("")
		return nil
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
	if err := d.Set("vm_provisioning", vlan.VMProvisioning); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	if len(vlan.Locations) > 0 {
		if err := d.Set("location_id", vlan.Locations[0].Identifier); err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
	}

	return diags
}

func resourceVLANUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(providerContext).legacyClient
	vlanAPI := vlan.NewAPI(c)

	if !d.HasChange("description_customer") && !d.HasChange("vm_provisioning") {
		return nil
	}

	def := vlan.UpdateDefinition{}
	if d.HasChange("description_customer") {
		def.CustomerDescription = d.Get("description_customer").(string)
	}
	def.VMProvisioning = d.Get("vm_provisioning").(bool)

	if err := vlanAPI.Update(ctx, d.Id(), def); err != nil {
		return diag.FromErr(err)
	}

	err := retry.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *retry.RetryError {
		vlanInfo, err := vlanAPI.Get(ctx, d.Id())
		if err != nil {
			if err := handleNotFoundError(err); err != nil {
				return retry.NonRetryableError(fmt.Errorf("unable to get vlan with id '%s': %w", d.Id(), err))
			}
			d.SetId("")
			return nil
		}
		if vlanInfo.VMProvisioning == d.Get("vm_provisioning").(bool) && vlanInfo.CustomerDescription == d.Get("description_customer").(string) {
			return nil
		}
		return retry.RetryableError(fmt.Errorf("waiting for vlan with id '%s' to be updated", d.Id()))
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceVLANRead(ctx, d, m)
}

func resourceVLANDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(providerContext).legacyClient
	v := vlan.NewAPI(c)

	err := v.Delete(ctx, d.Id())
	if err != nil {
		if err := handleNotFoundError(err); err != nil {
			return diag.FromErr(err)
		}
		d.SetId("")
		return nil
	}

	err = retry.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *retry.RetryError {
		info, err := v.Get(ctx, d.Id())
		if err != nil {
			if err := handleNotFoundError(err); err != nil {
				return retry.NonRetryableError(fmt.Errorf("unable to get vlan with id '%s': %w", d.Id(), err))
			}
			d.SetId("")
			return nil
		}
		if info.Status == vlanStatusDeleted {
			d.SetId("")
			return nil
		}
		return retry.RetryableError(fmt.Errorf("waiting for vlan with id '%s' to be deleted", d.Id()))
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
