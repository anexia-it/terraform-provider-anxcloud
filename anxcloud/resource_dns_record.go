package anxcloud

import (
	"context"
	"errors"
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
		CreateContext: resourceDNSRecordCreate,
		ReadContext:   resourceDNSRecordRead,
		UpdateContext: resourceDNSRecordUpdate,
		DeleteContext: resourceDNSRecordDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Read:   schema.DefaultTimeout(1 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: schemaDNSRecord(),
	}
}

func resourceDNSRecordCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	if v := ctx.Value(dnsRecordSkipMutexLock); v == nil {
		syncDNSRecordOps.Lock()
		defer syncDNSRecordOps.Unlock()
	}
	time.Sleep(time.Second)

	a := m.(providerContext).api

	r := dnsRecordFromResourceData(d)

	// try to import
	ret, err := findDNSRecord(ctx, a, r)
	if err != nil {
		if !errors.Is(err, api.ErrNotFound) {
			return diag.FromErr(err)
		}
		// no matching record found -> create new
		if err := a.Create(ctx, r); err != nil {
			return diag.FromErr(err)
		}
	} else {
		// record found -> import
		r = ret
	}

	d.SetId(r.Identifier)

	return resourceDNSRecordRead(ctx, d, m)
}

func resourceDNSRecordRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags []diag.Diagnostic

	a := m.(providerContext).api

	r := dnsRecordFromResourceData(d)
	r, err := findDNSRecord(ctx, a, r)
	if err != nil {
		if !errors.Is(err, api.ErrNotFound) {
			return diag.FromErr(err)
		}
		d.SetId("")
		return diags
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
	// remove quotes from txt rdata to prevent double, tripple, ... quoted data
	rData := r.RData
	if r.Type == "TXT" {
		rData = rData[1 : len(rData)-1]
	}
	if err := d.Set("rdata", rData); err != nil {
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
	if v := ctx.Value(dnsRecordSkipMutexLock); v == nil {
		syncDNSRecordOps.Lock()
		defer syncDNSRecordOps.Unlock()
	}
	time.Sleep(time.Second)

	a := m.(providerContext).api

	prevZoneName, _ := d.GetChange("zone_name")
	prevType, _ := d.GetChange("type")
	prevName, _ := d.GetChange("name")
	prevRData, _ := d.GetChange("rdata")
	prevTTL, _ := d.GetChange("ttl")

	r, err := findDNSRecord(ctx, a, &clouddnsv1.Record{
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
		if err := a.Destroy(ctx, &clouddnsv1.Record{Identifier: r.Identifier, ZoneName: prevZoneName.(string)}); err != nil {
			if !errors.Is(err, api.ErrNotFound) {
				return diag.FromErr(err)
			}
		}
		return resourceDNSRecordCreate(context.WithValue(ctx, dnsRecordSkipMutexLock, true), d, m)
	}

	revRecID := r.Identifier
	r = dnsRecordFromResourceData(d)
	r.Identifier = revRecID

	if err := a.Update(ctx, r); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(r.Identifier)

	return resourceDNSRecordRead(ctx, d, m)
}

func resourceDNSRecordDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	if v := ctx.Value(dnsRecordSkipMutexLock); v == nil {
		syncDNSRecordOps.Lock()
		defer syncDNSRecordOps.Unlock()
	}
	time.Sleep(time.Second)

	a := m.(providerContext).api

	r := dnsRecordFromResourceData(d)
	r, err := findDNSRecord(ctx, a, r)

	if err != nil {
		if !errors.Is(err, api.ErrNotFound) {
			return diag.FromErr(err)
		}
		d.SetId("")
		return nil
	}

	err = a.Destroy(ctx, r)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func dnsRecordFromResourceData(d *schema.ResourceData) *clouddnsv1.Record {
	return &clouddnsv1.Record{
		Type:      d.Get("type").(string),
		Name:      d.Get("name").(string),
		ZoneName:  d.Get("zone_name").(string),
		Region:    d.Get("region").(string),
		RData:     d.Get("rdata").(string),
		TTL:       d.Get("ttl").(int),
		Immutable: d.Get("immutable").(bool),
	}
}

func findDNSRecord(ctx context.Context, a api.API, r *clouddnsv1.Record) (*clouddnsv1.Record, error) {
	// quote TXTs rdata for compare.Compare
	if r.Type == "TXT" {
		r.RData = fmt.Sprintf("%q", r.RData)
	}

	channel := make(types.ObjectChannel)
	err := a.List(ctx, r, api.ObjectChannel(&channel))
	if err != nil {
		return nil, err
	}

	rec := &clouddnsv1.Record{}
	for res := range channel {
		err = res(rec)
		if err != nil {
			return nil, err
		}

		diffs, err := compare.Compare(r, rec, "Type", "Name", "RData", "TTL")
		if err != nil {
			return nil, err
		}

		if len(diffs) == 0 {
			return rec, nil
		}
	}

	return nil, api.ErrNotFound
}
