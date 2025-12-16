package anxcloud

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
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
			Create: schema.DefaultTimeout(30 * time.Minute),
			Read:   schema.DefaultTimeout(time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute), // Allow time for deployment waiting
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},
		Schema: schemaDNSZone(),
		CustomizeDiff: customdiff.All(
			validateZoneDoesNotExist,
		),
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

// waitForZoneDeployment waits for a zone to complete validation and deployment after updates
func waitForZoneDeployment(ctx context.Context, a api.API, zoneName string) error {
	// Add a brief initial delay to allow validation to start after update
	log.Printf("[DEBUG] DNS Zone Update: update executed, waiting for validation to start")
	time.Sleep(5 * time.Second)

	b := backoff.NewExponentialBackOff()
	b.InitialInterval = 10 * time.Second
	b.MaxInterval = 30 * time.Second
	b.MaxElapsedTime = 10 * time.Minute

	return backoff.Retry(func() error {
		zone := clouddnsv1.Zone{Name: zoneName}
		if err := a.Get(ctx, &zone); err != nil {
			return backoff.Permanent(err)
		}

		if zone.DeploymentLevel < 100 {
			return fmt.Errorf("waiting for zone deployment to complete: %d%%", zone.DeploymentLevel)
		}

		log.Printf("[DEBUG] DNS Zone Update: zone deployment complete (deployment: %d%%)", zone.DeploymentLevel)
		return nil
	}, backoff.WithContext(b, ctx))
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

	// Wait for zone update to complete deployment before returning
	// Zone updates trigger re-validation and re-deployment across all servers
	if err := waitForZoneDeployment(ctx, a, def.Name); err != nil {
		return diag.FromErr(fmt.Errorf("failed waiting for zone deployment after update: %w", err))
	}

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

// validateZoneDoesNotExist checks during plan if a zone with the same name already exists
//
// NOTE: This validation is intentionally commented out because it conflicts with
// import workflows. When using import blocks, the zone naturally already exists,
// and this validation would prevent the import from working.
//
// The original purpose was to warn users if they're about to create a zone that
// already exists. However, the resourceDNSZoneCreate function already handles this
// gracefully by adopting/updating existing zones, so the validation is redundant.
//
// If strict validation is needed in the future, it should detect import scenarios
// and skip validation, but Terraform's CustomizeDiff does not provide a reliable
// way to detect import operations.
func validateZoneDoesNotExist(ctx context.Context, d *schema.ResourceDiff, m interface{}) error {
	// Skip validation if this is an existing resource (update)
	if d.Id() != "" {
		return nil
	}

	// DISABLED: Validation conflicts with import workflows
	// See function documentation for details
	return nil

	/* Original validation code - DISABLED
	zoneName := d.Get("name").(string)

	// Only validate on create when name is being set
	if !d.HasChange("name") && d.Id() == "" {
		return nil
	}

	a := apiFromProviderConfig(m)
	log.Printf("[DEBUG] DNS Zone Plan: checking if zone '%s' already exists", zoneName)

	z := clouddnsv1.Zone{Name: zoneName}
	err := a.Get(ctx, &z)

	if api.IgnoreNotFound(err) != nil {
		// API error other than NotFound
		return fmt.Errorf("failed to check if zone exists: %w", err)
	}

	if err == nil {
		// Zone exists - warn user it will be adopted/updated instead of created
		log.Printf("[WARN] DNS Zone Plan: zone '%s' already exists and will be adopted/updated on apply", zoneName)
		return fmt.Errorf("DNS zone '%s' already exists. Use 'terraform import' to manage existing zones, or choose a different name", zoneName)
	}

	// Zone doesn't exist - safe to create
	log.Printf("[DEBUG] DNS Zone Plan: zone '%s' does not exist, will be created", zoneName)
	return nil
	*/
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
