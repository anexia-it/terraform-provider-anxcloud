package anxcloud

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"sort"
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

const (
	addressStateSuccess   = "success"
	addressStateReserving = "reserving address"
	addressStateWaitReady = "waiting for address to be ready"
	addressStateError     = "error"
)

func immediateErrorReturnRefreshFunc(err error) retry.StateRefreshFunc {
	return func() (any, string, error) {
		return nil, addressStateError, err
	}
}

func refreshAddressReadyState(ctx context.Context, addressClient address.API, id string, expectedStatus []string) (any, string, error) {
	addr, err := addressClient.Get(ctx, id)
	if err != nil {
		return nil, addressStateError, err
	} else if addr.VLANID == "" {
		return addr, addressStateWaitReady, nil
	}

	if len(expectedStatus) > 0 {
		sort.Strings(expectedStatus)

		if idx := sort.SearchStrings(expectedStatus, addr.Status); idx >= len(expectedStatus) || expectedStatus[idx] != addr.Status {
			return addr, addressStateWaitReady, nil
		}
	}

	return addr, addressStateSuccess, nil
}

func createAddress(ctx context.Context, pc providerContext, d *schema.ResourceData) retry.StateRefreshFunc {
	addressClient := address.NewAPI(pc.legacyClient)

	def := address.Create{
		Address:             d.Get("address").(string),
		PrefixID:            d.Get("network_prefix_id").(string),
		DescriptionCustomer: d.Get("description_customer").(string),
		Role:                d.Get("role").(string),
		Organization:        d.Get("organization").(string),
	}

	res, err := addressClient.Create(ctx, def)
	if err != nil {
		return immediateErrorReturnRefreshFunc(fmt.Errorf("error creating address: %w", err))
	}

	return func() (any, string, error) {
		return refreshAddressReadyState(ctx, addressClient, res.ID, []string{ipAddressStatusActive, ipAddressStatusInactive})
	}
}

func reserveAddress(ctx context.Context, pc providerContext, d *schema.ResourceData) retry.StateRefreshFunc {
	addressClient := address.NewAPI(pc.legacyClient)
	vlanClient := vlan.NewAPI(pc.legacyClient)

	vlan, err := vlanClient.Get(ctx, d.Get("vlan_id").(string))
	if err != nil {
		return immediateErrorReturnRefreshFunc(fmt.Errorf("error fetching vlan: %w", err))
	} else if len(vlan.Locations) < 1 {
		return immediateErrorReturnRefreshFunc(fmt.Errorf("vlan has no locations specified"))
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

	var successReserveSummary *address.ReserveRandomSummary

	return func() (any, string, error) {
		if successReserveSummary == nil {
			reserveSummary, err := addressClient.ReserveRandom(ctx, reserveOpts)

			var respErr *client.ResponseError
			if errors.As(err, &respErr) && respErr.ErrorData.Code == http.StatusConflict {
				// vlan or prefix might not be ready yet even though they are active
				return nil, addressStateReserving, nil
			} else if err != nil {
				return nil, addressStateError, fmt.Errorf("reserve endpoint returned an error: %w", err)
			} else if len(reserveSummary.Data) < 1 {
				return nil, addressStateError, fmt.Errorf("reserve endpoint didn't return any addresses")
			}

			successReserveSummary = &reserveSummary
		}

		return refreshAddressReadyState(ctx, addressClient, successReserveSummary.Data[0].ID, nil)
	}
}

func resourceIPAddressCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	wait := retry.StateChangeConf{
		Timeout:        d.Timeout(schema.TimeoutCreate),
		Target:         []string{addressStateSuccess},
		NotFoundChecks: math.MaxInt,
	}

	if _, ok := d.GetOk("address"); !ok {
		// if no `address` was provided, we will automatically reserve one
		wait.Refresh = reserveAddress(ctx, m.(providerContext), d)
		wait.Pending = []string{addressStateReserving, addressStateWaitReady}
	} else {
		// `address` was provided, so we create the specified address
		wait.Refresh = createAddress(ctx, m.(providerContext), d)
		wait.Pending = []string{addressStateWaitReady}
	}

	diags := make(diag.Diagnostics, 0)

	waitResult, err := wait.WaitForStateContext(ctx)
	if err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	if waitResult != nil {
		diags = append(diags, addressIntoResourceData(waitResult.(address.Address), d)...)
	}

	return diags
}

func addressIntoResourceData(a address.Address, d *schema.ResourceData) diag.Diagnostics {
	var diags []diag.Diagnostic

	d.SetId(a.ID)

	if err := d.Set("address", a.Name); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("network_prefix_id", a.PrefixID); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("status", a.Status); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("description_customer", a.DescriptionCustomer); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("description_internal", a.DescriptionInternal); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	// TODO: API require 'role' arg and returns 'role_text' arg, this must be fixed
	if err := d.Set("role", a.Role); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("version", a.Version); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("vlan_id", a.VLANID); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	return diags
}

func resourceIPAddressRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

	return addressIntoResourceData(info, d)
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
