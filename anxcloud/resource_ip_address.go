package anxcloud

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"go.anx.io/go-anxcloud/pkg/client"
	"go.anx.io/go-anxcloud/pkg/ipam/address"
	"go.anx.io/go-anxcloud/pkg/vlan"
)

const (
	ipAddressStatusInactive = "Inactive"
	ipAddressStatusActive   = "Active"
	ipAddressStatusDeleted  = "Marked for deletion"
)

func resourceIPAddress() *schema.Resource {
	return &schema.Resource{
		Description: "This resource allows you to create and configure IP addresses. " +
			"Addresses created without the `address` attribute will expire if the reservation period exceeds before assigned to a VM.",
		CreateContext: tagsMiddlewareCreate(resourceIPAddressCreate),
		ReadContext:   tagsMiddlewareRead(resourceIPAddressRead),
		UpdateContext: tagsMiddlewareUpdate(resourceIPAddressUpdate),
		DeleteContext: resourceIPAddressDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Read:   schema.DefaultTimeout(1 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: withTagsAttribute(
			map[string]*schema.Schema{
				"id": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: identifierDescription,
				},
				"address": {
					Type:         schema.TypeString,
					Optional:     true,
					Computed:     true,
					ForceNew:     true,
					Description:  "IP address.",
					ExactlyOneOf: []string{"address", "vlan_id"},
					RequiredWith: []string{"network_prefix_id"}, // network_prefix_id is required if address is set
				},
				"vlan_id": {
					Type:         schema.TypeString,
					Optional:     true,
					Computed:     true,
					ForceNew:     true,
					Description:  "The associated VLAN identifier.",
					ExactlyOneOf: []string{"address", "vlan_id"},
				},
				"network_prefix_id": {
					Type:        schema.TypeString,
					Optional:    true,
					Computed:    true,
					ForceNew:    true,
					Description: "Identifier of the related network prefix.",
				},
				"version": {
					Type:        schema.TypeInt,
					Optional:    true,
					Computed:    true,
					ForceNew:    true,
					Description: "IP version.",
				},
				"description_customer": {
					Type:          schema.TypeString,
					Optional:      true,
					Computed:      true,
					Description:   "Additional customer description.",
					ConflictsWith: []string{"vlan_id"},
				},
				"description_internal": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Internal description.",
				},
				"role": {
					Type:          schema.TypeString,
					Optional:      true,
					Default:       "Default",
					Description:   "Role of the IP address",
					ConflictsWith: []string{"vlan_id"},
				},
				"organization": {
					Type:          schema.TypeString,
					Optional:      true,
					Computed:      true,
					Description:   "Customer of yours. Reseller only.",
					ConflictsWith: []string{"vlan_id"},
				},
				"status": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Status of the IP address",
				},
				"reservation_period_seconds": {
					Type:          schema.TypeInt,
					Optional:      true,
					ConflictsWith: []string{"address"},
					Description:   "Period for the requested reservation in seconds. Defaults to 30 minutes if not set.",
				},
			},
		),
	}
}

func resourceIPAddressCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(providerContext).legacyClient
	a := address.NewAPI(c)
	v := vlan.NewAPI(c)

	// if `vlan_id` was provided, we will perform an ip reservation
	if vlanID, ok := d.GetOk("vlan_id"); ok {
		vlan, err := v.Get(ctx, vlanID.(string))
		if err != nil {
			return diag.FromErr(fmt.Errorf("fetch vlan: %w", err))
		} else if len(vlan.Locations) < 1 {
			return diag.Errorf("vlan has no locations specified")
		}

		reserveOpts := address.ReserveRandom{
			VlanID:     vlan.Identifier,
			LocationID: vlan.Locations[0].Identifier,
			Count:      1,
		}

		if reservationPeriodSeconds, ok := d.GetOk("reservation_period_seconds"); ok {
			reserveOpts.ReservationPeriod = uint(reservationPeriodSeconds.(int))
		}

		if prefixID, ok := d.GetOk("network_prefix_id"); ok {
			reserveOpts.PrefixID = prefixID.(string)
		} else if ipVersion, ok := d.GetOk("version"); ok {
			reserveOpts.IPVersion = address.IPReserveVersionLimit(ipVersion.(int))
		}

		var reserveSummary address.ReserveRandomSummary
		if err := retry.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *retry.RetryError {
			var (
				err     error
				respErr *client.ResponseError
			)

			if reserveSummary, err = a.ReserveRandom(ctx, reserveOpts); errors.As(err, &respErr) && respErr.ErrorData.Code == http.StatusConflict {
				// vlan or prefix might not be ready yet even though they are active
				return retry.RetryableError(err)
			} else if err != nil {
				return retry.NonRetryableError(fmt.Errorf("reserve address: %w", err))
			} else if len(reserveSummary.Data) < 1 {
				return retry.NonRetryableError(fmt.Errorf("reserve endpoint didn't return any addresses"))
			}

			return nil
		}); err != nil {
			return diag.FromErr(err)
		}

		d.SetId(reserveSummary.Data[0].ID)

		if err := retry.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *retry.RetryError {
			if addr, err := a.Get(ctx, reserveSummary.Data[0].ID); err != nil {
				return retry.NonRetryableError(err)
			} else if addr.VLANID == "" {
				return retry.RetryableError(fmt.Errorf("VLAN id is not set"))
			}

			return nil
		}); err != nil {
			return diag.FromErr(fmt.Errorf("wait for VLAN to be set on address resource: %w", err))
		}

		return resourceIPAddressRead(ctx, d, m)
	}

	// create specific address
	def := address.Create{
		Address:             d.Get("address").(string),
		PrefixID:            d.Get("network_prefix_id").(string),
		DescriptionCustomer: d.Get("description_customer").(string),
		Role:                d.Get("role").(string),
		Organization:        d.Get("organization").(string),
	}
	res, err := a.Create(ctx, def)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(res.ID)

	err = retry.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *retry.RetryError {
		info, err := a.Get(ctx, d.Id())
		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("unable to get ip address with '%s' id: %w", d.Id(), err))
		}
		if info.Status == ipAddressStatusInactive || info.Status == ipAddressStatusActive {
			return nil
		}
		return retry.RetryableError(fmt.Errorf("waiting for ip address with '%s' id to be ready", d.Id()))
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceIPAddressRead(ctx, d, m)
}

func resourceIPAddressRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags []diag.Diagnostic

	c := m.(providerContext).legacyClient
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
	c := m.(providerContext).legacyClient
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
	c := m.(providerContext).legacyClient
	a := address.NewAPI(c)

	if addr, err := a.Get(ctx, d.Id()); isLegacyNotFoundError(err) {
		// handle not found error by just deleting the resource
		d.SetId("")
		return nil
	} else if err != nil {
		// return unhandled error
		return diag.FromErr(err)
	} else if addr.DescriptionInternal == "reserved" {
		d.SetId("")
		var diags diag.Diagnostics
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Could not delete reserved address",
			Detail:   "Reserved addresses cannot be deleted manually. They'll expire eventually.",
		})
		return diags
	}

	err := a.Delete(ctx, d.Id())
	if err != nil {
		if err := handleNotFoundError(err); err != nil {
			return diag.FromErr(err)
		}
		d.SetId("")
		return nil
	}

	err = retry.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *retry.RetryError {
		info, err := a.Get(ctx, d.Id())
		if err != nil {
			if err := handleNotFoundError(err); err != nil {
				return retry.NonRetryableError(fmt.Errorf("unable to get ip address with id '%s': %w", d.Id(), err))
			}
			d.SetId("")
			return nil
		}
		if info.Status == ipAddressStatusDeleted {
			d.SetId("")
			return nil
		}
		return retry.RetryableError(fmt.Errorf("waiting for ip address with id '%s' to be deleted", d.Id()))
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
