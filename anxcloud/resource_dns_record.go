package anxcloud

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"go.anx.io/go-anxcloud/pkg/api"
	"go.anx.io/go-anxcloud/pkg/api/types"
	clouddnsv1 "go.anx.io/go-anxcloud/pkg/apis/clouddns/v1"
	"go.anx.io/go-anxcloud/pkg/utils/object/compare"
)

var syncDNSRecordOps sync.Mutex
var dnsRecordSkipMutexLock = providerContextKey("dns-record-skip-mutex-lock")

func resourceDNSRecord() *schema.Resource {
	return &schema.Resource{
		Description:   "This resource allows you to create DNS records for a specified zone. TXT records might behave funny, we are working on it.",
		CreateContext: resourceDNSRecordCreate,
		ReadContext:   resourceDNSRecordRead,
		UpdateContext: resourceDNSRecordUpdate,
		DeleteContext: resourceDNSRecordDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(time.Minute),
			Read:   schema.DefaultTimeout(time.Minute),
			Update: schema.DefaultTimeout(time.Minute),
			Delete: schema.DefaultTimeout(time.Minute),
		},
		Schema: schemaDNSRecord(),
	}
}

func resourceDNSRecordCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	defer lockRecordContextIfNeeded(ctx)()

	a := apiFromProviderConfig(m)

	r := dnsRecordFromResourceData(d)

	// try to import
	if ret, err := findDNSRecord(ctx, a, r); api.IgnoreNotFound(err) != nil {
		return diag.FromErr(err)
	} else if err == nil {
		// DNS Record found
		d.SetId(ret.Identifier)
		return resourceDNSRecordRead(ctx, d, m)
	}

	// not found -> create new zone

	if err := a.Create(ctx, &r); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(r.Identifier)

	return resourceDNSRecordRead(ctx, d, m)
}

func resourceDNSRecordRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags []diag.Diagnostic

	a := apiFromProviderConfig(m)

	r := dnsRecordFromResourceData(d)
	r, err := findDNSRecord(ctx, a, r)

	if api.IgnoreNotFound(err) != nil {
		return diag.FromErr(err)
	} else if err != nil {
		d.SetId("")
		return diags
	}

	// remove quotes from txt rdata to prevent double, tripple, ... quoted data (SYSENG-816)
	rData := r.RData
	if r.Type == "TXT" {
		rData = rData[1 : len(rData)-1]
	}

	if err := d.Set("rdata", rData); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("type", r.Type); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("name", r.Name); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("zone_name", r.ZoneName); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("region", r.Region); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("ttl", r.TTL); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("immutable", r.Immutable); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	return diags
}

func resourceDNSRecordUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	defer lockRecordContextIfNeeded(ctx)()

	a := apiFromProviderConfig(m)

	prevZoneName, _ := d.GetChange("zone_name")
	prevType, _ := d.GetChange("type")
	prevName, _ := d.GetChange("name")
	prevRData, _ := d.GetChange("rdata")
	prevTTL, _ := d.GetChange("ttl")

	r, err := findDNSRecord(ctx, a, clouddnsv1.Record{
		ZoneName: prevZoneName.(string),
		Type:     prevType.(string),
		Name:     prevName.(string),
		RData:    prevRData.(string),
		TTL:      prevTTL.(int),
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("zone_name") { // cannot change records zone -> recreate
		err := a.Destroy(ctx, &clouddnsv1.Record{Identifier: r.Identifier, ZoneName: prevZoneName.(string)})
		if api.IgnoreNotFound(err) != nil {
			return diag.FromErr(err)
		}
		return resourceDNSRecordCreate(context.WithValue(ctx, dnsRecordSkipMutexLock, true), d, m)
	}

	revRecID := r.Identifier
	r = dnsRecordFromResourceData(d)
	r.Identifier = revRecID

	if err := a.Update(ctx, &r); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(r.Identifier)

	return resourceDNSRecordRead(ctx, d, m)
}

func resourceDNSRecordDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	defer lockRecordContextIfNeeded(ctx)()

	a := apiFromProviderConfig(m)

	r := dnsRecordFromResourceData(d)
	r, err := findDNSRecord(ctx, a, r)

	if api.IgnoreNotFound(err) != nil {
		return diag.FromErr(err)
	} else if err != nil {
		d.SetId("")
		return nil
	}

	err = a.Destroy(ctx, &r)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func dnsRecordFromResourceData(d *schema.ResourceData) clouddnsv1.Record {
	return clouddnsv1.Record{
		Type:      d.Get("type").(string),
		Name:      d.Get("name").(string),
		ZoneName:  d.Get("zone_name").(string),
		Region:    d.Get("region").(string),
		RData:     d.Get("rdata").(string),
		TTL:       d.Get("ttl").(int),
		Immutable: d.Get("immutable").(bool),
	}
}

func findDNSRecord(ctx context.Context, a api.API, r clouddnsv1.Record) (foundRecord clouddnsv1.Record, err error) {
	// quote TXTs rdata for compare.Compare (SYSENG-816)
	if r.Type == "TXT" {
		r.RData = fmt.Sprintf("%q", r.RData)
	}

	var pageIter types.PageInfo
	err = a.List(ctx, &r, api.Paged(1, 100, &pageIter))
	if err != nil {
		return
	}

	var pagedRecords []clouddnsv1.Record
	for pageIter.Next(&pagedRecords) {
		idx, err := compare.Search(&r, pagedRecords, "Type", "Name", "RData", "TTL")
		if err != nil {
			return foundRecord, err
		}
		if idx > -1 {
			return pagedRecords[idx], nil
		}
	}

	return foundRecord, api.ErrNotFound
}

func lockRecordContextIfNeeded(ctx context.Context) func() {
	if v := ctx.Value(dnsRecordSkipMutexLock); v == nil {
		syncDNSRecordOps.Lock()
		return syncDNSRecordOps.Unlock
	}
	return func() { /* noop */ }
}