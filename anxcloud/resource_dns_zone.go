package anxcloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"go.anx.io/go-anxcloud/pkg/api"
	clouddnsv1 "go.anx.io/go-anxcloud/pkg/apis/clouddns/v1"
)

func resourceDNSZone() *schema.Resource {
	return &schema.Resource{
		Description:   "This resource allows you to create DNS zones.",
		CreateContext: resourceDNSZoneCreate,
		ReadContext:   resourceDNSZoneRead,
		UpdateContext: resourceDNSZoneUpdate,
		DeleteContext: resourceDNSZoneDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(time.Minute),
			Read:   schema.DefaultTimeout(time.Minute),
			Update: schema.DefaultTimeout(time.Minute),
			Delete: schema.DefaultTimeout(time.Minute),
		},
		Schema: schemaDNSZone(),
	}
}

func resourceDNSZoneCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	a := apiFromProviderConfig(m)

	// try to import
	z := clouddnsv1.Zone{Name: d.Get("name").(string)}

	if err := a.Get(ctx, &z); api.IgnoreNotFound(err) != nil {
		return diag.FromErr(err)
	} else if err == nil {
		// DNS Zone found -> update to match terraform definition
		d.SetId(z.Name)
		return resourceDNSZoneUpdate(ctx, d, m)
	}

	// not found -> create new zone

	z = dnsZoneFromResourceData(d)
	if err := a.Create(ctx, &z); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(z.Name)

	return resourceDNSZoneRead(ctx, d, m)
}

func resourceDNSZoneRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags []diag.Diagnostic

	a := apiFromProviderConfig(m)

	z := clouddnsv1.Zone{Name: d.Id()}

	err := a.Get(ctx, &z)

	if api.IgnoreNotFound(err) != nil {
		return diag.FromErr(err)
	} else if err != nil {
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
	a := apiFromProviderConfig(m)

	if d.HasChange("name") && !d.IsNewResource() {
		return diag.FromErr(fmt.Errorf("%w: cannot change the name of a DNS zone", ErrOperationNotSupported))
	}

	def := dnsZoneFromResourceData(d)

	if err := a.Update(ctx, &def); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(def.Name)

	return resourceDNSZoneRead(ctx, d, m)
}

func resourceDNSZoneDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	a := apiFromProviderConfig(m)

	z := clouddnsv1.Zone{Name: d.Id()}

	err := a.Destroy(ctx, &z)
	if api.IgnoreNotFound(err) != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func dnsZoneFromResourceData(d *schema.ResourceData) clouddnsv1.Zone {
	dnsServers := expandDNSServers(d.Get("dns_servers").([]interface{}))

	notifyAllowedIPsAsInterfaces := d.Get("notify_allowed_ips").([]interface{})
	notifyAllowedIPs := make([]string, 0, len(notifyAllowedIPsAsInterfaces))
	for _, v := range notifyAllowedIPsAsInterfaces {
		notifyAllowedIPs = append(notifyAllowedIPs, v.(string))
	}

	return clouddnsv1.Zone{
		Name:             d.Get("name").(string),
		IsMaster:         d.Get("is_master").(bool),
		DNSSecMode:       d.Get("dns_sec_mode").(string),
		AdminEmail:       d.Get("admin_email").(string),
		Refresh:          d.Get("refresh").(int),
		Retry:            d.Get("retry").(int),
		Expire:           d.Get("expire").(int),
		TTL:              d.Get("ttl").(int),
		MasterNS:         d.Get("master_nameserver").(string),
		DNSServers:       dnsServers,
		NotifyAllowedIPs: notifyAllowedIPs,
	}
}
