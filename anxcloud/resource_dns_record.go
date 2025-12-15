package anxcloud

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"go.anx.io/go-anxcloud/pkg/api"
	clouddnsv1 "go.anx.io/go-anxcloud/pkg/apis/clouddns/v1"
)

func resourceDNSRecord() *schema.Resource {
	return &schema.Resource{
		Description: "This resource allows you to create DNS records for a specified zone. TXT records might behave funny, we are working on it." +
			" Create and delete operations will be handled in batches internally. As a side effect this will cause whole batches to fail in case some of the operations are invalid." +
			" TTL and RDATA fields can be updated in-place without requiring record replacement.",
		CreateContext: resourceDNSRecordCreate,
		ReadContext:   resourceDNSRecordRead,
		UpdateContext: resourceDNSRecordUpdate,
		DeleteContext: resourceDNSRecordDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceDNSRecordImport,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Read:   schema.DefaultTimeout(time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},
		Schema: schemaDNSRecord(),
	}
}

var resourceDNSRecordBatcherMap sync.Map

func resourceDNSRecordCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	a := apiFromProviderConfig(m)
	batcher := resourceDNSRecordBatcherForZone(a, d.Get("zone_name").(string))

	r := dnsRecordFromResourceData(d)

	// For TXT records, add DNS protocol quotes (SYSENG-816)
	// DNS TXT records must be quoted per RFC
	// We add quotes here so users don't have to worry about it in their config
	// We manually add quotes instead of using %q to preserve user's quote characters
	if r.Type == "TXT" {
		r.RData = `"` + r.RData + `"`
	}

	// Debug logging
	log.Printf("[DEBUG] DNS Record Create: zone=%s, name=%s, type=%s, rdata=%s, ttl=%d", r.ZoneName, r.Name, r.Type, r.RData, r.TTL)

	// try to import
	if _, err := findDNSRecord(ctx, a, r); api.IgnoreNotFound(err) != nil {
		return diag.FromErr(err)
	} else if err == nil {
		// DNS Record found - generate content hash for existing record
		contentID := generateContentHash(r.ZoneName, r.Name, r.Type, r.RData, r.TTL)
		d.SetId(contentID)
		return resourceDNSRecordRead(ctx, d, m)
	}

	// not found -> create new record
	// Generate content hash immediately for stable Terraform resource ID
	contentID := generateContentHash(r.ZoneName, r.Name, r.Type, r.RData, r.TTL)
	d.SetId(contentID)

	if _, err := batcher.Process(ctx, recordBatchUnit{record: r, batchOperation: batchOperationCreate}); err != nil {
		return diag.FromErr(err)
	}

	// Wait for zone deployment to complete before verification using coordinated polling
	// This ensures records are fully propagated and available for reading
	// The coordinator manages polling to avoid redundant API calls during concurrent operations
	log.Printf("[DEBUG] DNS Record Create: changeset executed, coordinating zone deployment polling")

	coordinator := getZonePollingCoordinator(a, r.ZoneName)
	defer coordinator.release()

	if deploymentErr := coordinator.waitForZoneDeployment(ctx); deploymentErr != nil {
		return diag.FromErr(fmt.Errorf("failed waiting for zone deployment: %w", deploymentErr))
	}

	return resourceDNSRecordRead(ctx, d, m)
}

func resourceDNSRecordRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags []diag.Diagnostic

	a := apiFromProviderConfig(m)

	// Get target record data
	targetRecord := dnsRecordFromResourceData(d)

	log.Printf("[DEBUG] DNS Record Read: id=%s, zone=%s, name=%s, type=%s, rdata=%s, ttl=%d", d.Id(), targetRecord.ZoneName, targetRecord.Name, targetRecord.Type, targetRecord.RData, targetRecord.TTL)

	// Fetch all zone records to get current real identifiers
	zoneRecords, err := fetchAllZoneRecords(ctx, a, targetRecord.ZoneName)
	if err != nil {
		return diag.FromErr(err)
	}

	// First try to find by identifier if we have one (for imported/updated records)
	var realRecord *clouddnsv1.Record
	if d.Id() != "" {
		log.Printf("[DEBUG] DNS Record Read: trying to find record by identifier %s among %d zone records", d.Id(), len(zoneRecords))
		for _, record := range zoneRecords {
			log.Printf("[DEBUG] DNS Record Read: checking record identifier %s", record.Identifier)
			if record.Identifier == d.Id() {
				realRecord = &record
				log.Printf("[DEBUG] DNS Record Read: found record by identifier match")
				break
			}
		}
		if realRecord == nil {
			log.Printf("[DEBUG] DNS Record Read: record with identifier %s not found in zone, will try content matching", d.Id())
		}
	}

	// If not found by identifier, try content matching
	if realRecord == nil {
		// Determine if we should ignore TTL and region in matching
		// TTL: Ignore if it's 0 and not explicitly set (API uses zone default)
		ignoreTTL := targetRecord.TTL == 0
		// Region: Always ignore since it's computed and may not match API values during read
		ignoreRegion := true

		foundRecord, err := findRecordByContentFlexible(zoneRecords, targetRecord, ignoreTTL, ignoreRegion)
		if err != nil {
			// Record not found
			d.SetId("")
			return diags
		}
		realRecord = foundRecord
	}

	// For TXT records, strip the DNS protocol quotes (SYSENG-816)
	// DNS TXT records are always quoted per RFC, so API returns them with quotes
	// We strip the outer quotes to make it user-friendly in Terraform config
	// Users who want literal quotes in the TXT value should escape them in HCL
	rData := realRecord.RData
	if realRecord.Type == "TXT" {
		// Safely strip outer quotes if present
		rData = stripOuterQuotes(rData)
	}

	// Use the stable backend identifier as the resource ID
	d.SetId(realRecord.Identifier)

	if err := d.Set("identifier", realRecord.Identifier); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("rdata", rData); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("type", realRecord.Type); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("name", realRecord.Name); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("zone_name", realRecord.ZoneName); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("region", realRecord.Region); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("ttl", realRecord.TTL); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("immutable", realRecord.Immutable); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	return diags
}

func resourceDNSRecordDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	a := apiFromProviderConfig(m)
	batcher := resourceDNSRecordBatcherForZone(a, d.Get("zone_name").(string))

	// Get target record data
	targetRecord := dnsRecordFromResourceData(d)

	// Fetch all zone records to get current real identifiers
	zoneRecords, err := fetchAllZoneRecords(ctx, a, targetRecord.ZoneName)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to fetch zone records: %w", err))
	}

	// Find the matching record by content with flexible matching
	ignoreTTL := targetRecord.TTL == 0
	ignoreRegion := true
	realRecord, err := findRecordByContentFlexible(zoneRecords, targetRecord, ignoreTTL, ignoreRegion)
	if err != nil {
		// Record not found - already deleted
		d.SetId("")
		return nil
	}

	// Delete using the real record with current identifier
	if _, err := batcher.Process(ctx, recordBatchUnit{record: *realRecord, batchOperation: batchOperationDelete}); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func resourceDNSRecordUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags []diag.Diagnostic

	a := apiFromProviderConfig(m)

	log.Printf("[DEBUG] DNS Record Update: id=%s, zone=%s", d.Id(), d.Get("zone_name").(string))

	// Check if record is immutable
	if immutable := d.Get("immutable").(bool); immutable {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Cannot update immutable DNS record",
			Detail:   "This DNS record is marked as immutable and cannot be modified. To change this record, you must delete and recreate it.",
		})
		return diags
	}

	// Use the stored backend identifier to find the record to update
	storedIdentifier := d.Id()
	if storedIdentifier == "" {
		return diag.FromErr(fmt.Errorf("resource ID is empty"))
	}

	log.Printf("[DEBUG] DNS Record Update: attempting direct lookup by identifier=%s", storedIdentifier)

	// Try to get the record directly by identifier
	targetRecord := clouddnsv1.Record{
		Identifier: storedIdentifier,
		ZoneName:   d.Get("zone_name").(string),
	}

	realRecord, err := findDNSRecord(ctx, a, targetRecord)
	if err != nil {
		log.Printf("[DEBUG] DNS Record Update: direct lookup failed for id=%s, falling back to content matching", storedIdentifier)

		// If direct lookup fails, fall back to content matching
		// Get target record data - start with current configuration values
		targetRecord = dnsRecordFromResourceData(d)

		// For content matching, use OLD values for mutable fields (rdata, ttl) to find existing record
		// The API still has the old values, so we need to search using those
		if oldRdata, newRdata := d.GetChange("rdata"); oldRdata != newRdata {
			targetRecord.RData = oldRdata.(string) // Use old rdata to find existing record
		}
		if oldTTL, newTTL := d.GetChange("ttl"); oldTTL != newTTL {
			targetRecord.TTL = oldTTL.(int) // Use old ttl to find existing record
		}

		// Fetch all zone records to get current real identifiers
		zoneRecords, err := fetchAllZoneRecords(ctx, a, targetRecord.ZoneName)
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to fetch zone records: %w", err))
		}

		// Find the matching record by content using old values for mutable fields
		// Use same flexible matching as read operation
		ignoreTTL := targetRecord.TTL == 0
		ignoreRegion := true
		foundRecord, err := findRecordByContentFlexible(zoneRecords, targetRecord, ignoreTTL, ignoreRegion)
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to find record to update: %w", err))
		}
		realRecord = *foundRecord
		log.Printf("[DEBUG] DNS Record Update: found record by content matching, identifier=%s", realRecord.Identifier)
	} else {
		log.Printf("[DEBUG] DNS Record Update: direct lookup succeeded, found record identifier=%s", realRecord.Identifier)
	}

	// Prepare update record with minimal payload - only updatable fields
	// Exclude computed fields (Region, Immutable) and immutable fields (Type, Name, ZoneName)
	// Include Identifier for API routing and updatable fields (RData, TTL)

	// Get new RData value from state
	updateRData := d.Get("rdata").(string)

	// For TXT records, add DNS protocol quotes
	// We manually add quotes instead of using %q to preserve user's quote characters
	if realRecord.Type == "TXT" {
		updateRData = `"` + updateRData + `"`
	}

	updateRecord := clouddnsv1.Record{
		Identifier: realRecord.Identifier, // Required for API routing
		RData:      updateRData,           // New RData value (quoted for TXT)
		TTL:        d.Get("ttl").(int),    // New TTL value
	}

	// Use the CloudDNS API Update method
	if err := a.Update(ctx, &updateRecord); err != nil {
		return diag.FromErr(fmt.Errorf("failed to update DNS record: %w", err))
	}

	return resourceDNSRecordRead(ctx, d, m)
}

func resourceDNSRecordImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	importID := d.Id()

	// Support both old format (<zone_name>/<uuid>) and new format (<zone_name>/<content_hash>)
	parts := strings.Split(importID, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid import ID format: expected '<zone_name>/<record_identifier>', got '%s'. If importing existing DNS records, consider using the generate-import-blocks.sh script to generate proper import commands", importID)
	}

	zoneName := parts[0]
	recordIdentifier := parts[1]

	// Validate UUID format if it appears to be a UUID
	uuidRegex := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	if strings.Contains(recordIdentifier, "-") && !uuidRegex.MatchString(recordIdentifier) {
		return nil, fmt.Errorf("invalid UUID format in import ID: '%s'. UUIDs must be in format xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx", recordIdentifier)
	}

	// Set the zone name in the resource data
	if err := d.Set("zone_name", zoneName); err != nil {
		return nil, fmt.Errorf("failed to set zone_name: %w", err)
	}

	a := apiFromProviderConfig(m)

	// First, try to find the record by its real API identifier (works for both old and new imports)
	targetRecord := clouddnsv1.Record{
		Identifier: recordIdentifier,
		ZoneName:   zoneName,
	}

	foundRecord, err := findDNSRecord(ctx, a, targetRecord)
	if err != nil {
		// If not found by identifier, this might be a content hash or fake UUID
		// Log warning on fallback
		log.Printf("[WARN] DNS record import: identifier '%s' not found in zone '%s', falling back to content-based lookup", recordIdentifier, zoneName)
		// Set the ID and let the read operation handle finding by content
		d.SetId(recordIdentifier)
		return []*schema.ResourceData{d}, nil
	}

	// Found by identifier - set all the fields from the record
	if err := d.Set("name", foundRecord.Name); err != nil {
		return nil, fmt.Errorf("failed to set name: %w", err)
	}
	if err := d.Set("type", foundRecord.Type); err != nil {
		return nil, fmt.Errorf("failed to set type: %w", err)
	}
	// For TXT records, strip DNS protocol quotes (same as Read function)
	// This ensures consistency between import and normal read operations
	rdata := foundRecord.RData
	if foundRecord.Type == "TXT" {
		rdata = stripOuterQuotes(rdata)
	}

	if err := d.Set("rdata", rdata); err != nil {
		return nil, fmt.Errorf("failed to set rdata: %w", err)
	}
	if err := d.Set("ttl", foundRecord.TTL); err != nil {
		return nil, fmt.Errorf("failed to set ttl: %w", err)
	}
	if err := d.Set("region", foundRecord.Region); err != nil {
		return nil, fmt.Errorf("failed to set region: %w", err)
	}
	if err := d.Set("identifier", foundRecord.Identifier); err != nil {
		return nil, fmt.Errorf("failed to set identifier: %w", err)
	}
	if err := d.Set("immutable", foundRecord.Immutable); err != nil {
		return nil, fmt.Errorf("failed to set immutable: %w", err)
	}

	// Use the stable backend identifier as the resource ID
	d.SetId(foundRecord.Identifier)

	return []*schema.ResourceData{d}, nil
}
