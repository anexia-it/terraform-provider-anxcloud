package anxcloud

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"go.anx.io/go-anxcloud/pkg/api"
	clouddnsv1 "go.anx.io/go-anxcloud/pkg/apis/clouddns/v1"
)

func resourceDNSZone() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDNSZoneCreate,
		ReadContext:   resourceDNSZoneRead,
		UpdateContext: resourceDNSZoneUpdate,
		DeleteContext: resourceDNSZoneDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Read:   schema.DefaultTimeout(1 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: schemaDNSZone(),
	}
}

func resourceDNSZoneCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	a := m.(providerContext).api

	// try to import
	z := &clouddnsv1.Zone{Name: d.Get("name").(string)}
	if err := a.Get(ctx, z); err != nil {
		if !errors.Is(err, api.ErrNotFound) {
			return diag.FromErr(err)
		}
		// not found -> create new zone
	} else {
		// DNS Zone found -> update to match terraform definition
		d.SetId(z.Name)
		return resourceDNSZoneUpdate(ctx, d, m)
	}

	z = dnsZoneFromResourceData(d)
	if err := a.Create(ctx, z); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(z.Name)

	return resourceDNSZoneRead(ctx, d, m)
}

func resourceDNSZoneRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags []diag.Diagnostic

	a := m.(providerContext).api

	z := clouddnsv1.Zone{Name: d.Id()}

	err := a.Get(ctx, &z)

	if err != nil {
		if !errors.Is(err, api.ErrNotFound) {
			return diag.FromErr(err)
		}
		d.SetId("")
		return nil
	}

	if err := d.Set("name", z.Name); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("is_master", z.IsMaster); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("dns_sec_mode", z.DNSSecMode); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("admin_email", z.AdminEmail); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("refresh", z.Refresh); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("retry", z.Retry); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("expire", z.Expire); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("ttl", z.TTL); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("master_nameserver", z.MasterNS); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("notify_allowed_ips", z.NotifyAllowedIPs); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	flattenedDNSServers := flattenDNSServers(z.DNSServers)
	if err := d.Set("dns_servers", flattenedDNSServers); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("is_editable", z.IsEditable); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("validation_level", z.ValidationLevel); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("deployment_level", z.DeploymentLevel); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	return diags
}

func resourceDNSZoneUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	a := m.(providerContext).api

	if d.HasChange("name") {
		return diag.FromErr(fmt.Errorf("%w: cannot change the name of a DNS zone", ErrOperationNotSupported))
	}

	def := dnsZoneFromResourceData(d)

	if err := a.Update(ctx, def); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(def.Name)

	return resourceDNSZoneRead(ctx, d, m)
}

func resourceDNSZoneDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	a := m.(providerContext).api

	z := clouddnsv1.Zone{Name: d.Id()}

	err := a.Destroy(ctx, &z)
	if err != nil {
		if !errors.Is(err, api.ErrNotFound) {
			return diag.FromErr(err)
		}
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func dnsZoneFromResourceData(d *schema.ResourceData) *clouddnsv1.Zone {
	var notifyAllowedIPs []string
	for _, v := range d.Get("notify_allowed_ips").([]interface{}) {
		notifyAllowedIPs = append(notifyAllowedIPs, v.(string))
	}

	return &clouddnsv1.Zone{
		Name:             d.Get("name").(string),
		IsMaster:         d.Get("is_master").(bool),
		DNSSecMode:       d.Get("dns_sec_mode").(string),
		AdminEmail:       d.Get("admin_email").(string),
		Refresh:          d.Get("refresh").(int),
		Retry:            d.Get("retry").(int),
		Expire:           d.Get("expire").(int),
		TTL:              d.Get("ttl").(int),
		MasterNS:         d.Get("master_nameserver").(string),
		DNSServers:       expandDNSServers(d.Get("dns_servers").([]interface{})),
		NotifyAllowedIPs: notifyAllowedIPs,
	}
}
