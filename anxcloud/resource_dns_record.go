package anxcloud

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/anexia-it/terraform-provider-anxcloud/anxcloud/internal/utils"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"go.anx.io/go-anxcloud/pkg/api"
	"go.anx.io/go-anxcloud/pkg/api/types"
	clouddnsv1 "go.anx.io/go-anxcloud/pkg/apis/clouddns/v1"
	"go.anx.io/go-anxcloud/pkg/utils/object/compare"
)

func resourceDNSRecord() *schema.Resource {
	return &schema.Resource{
		Description: "This resource allows you to create DNS records for a specified zone. TXT records might behave funny, we are working on it." +
			" Create and delete operations will be handled in batches internally. As a side effect this will cause whole batches to fail in case some of the operations are invalid." +
			" Updating record attributes triggers a replacement (destroy old -> create new).",
		CreateContext: resourceDNSRecordCreate,
		ReadContext:   resourceDNSRecordRead,
		DeleteContext: resourceDNSRecordDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(2 * time.Minute),
			Read:   schema.DefaultTimeout(time.Minute),
			Delete: schema.DefaultTimeout(2 * time.Minute),
		},
		Schema: schemaDNSRecord(),
	}
}

var resourceDNSRecordBatcherMap sync.Map

func resourceDNSRecordBatcherForZone(a api.API, zoneName string) *utils.Batcher[recordBatchUnit, any] {
	anyBatcher, _ := resourceDNSRecordBatcherMap.LoadOrStore(zoneName, &utils.Batcher[recordBatchUnit, any]{
		// this will consume 15 seconds of the 2 minute create/delete budget
		Wait:      15 * time.Second,
		BatchFunc: resourceDNSRecordBatch(a, zoneName),
	})

	return anyBatcher.(*utils.Batcher[recordBatchUnit, any])
}

func resourceDNSRecordCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	a := apiFromProviderConfig(m)
	batcher := resourceDNSRecordBatcherForZone(a, d.Get("zone_name").(string))

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

	if _, err := batcher.Process(ctx, recordBatchUnit{record: r, batchOperation: batchOperationCreate}); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resourceDNSRecordCanonicalIdentifier(r))

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

func resourceDNSRecordDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	a := apiFromProviderConfig(m)
	batcher := resourceDNSRecordBatcherForZone(a, d.Get("zone_name").(string))

	r := dnsRecordFromResourceData(d)
	r, err := findDNSRecord(ctx, a, r)

	if api.IgnoreNotFound(err) != nil {
		return diag.FromErr(err)
	} else if err != nil {
		d.SetId("")
		return nil
	}

	if _, err := batcher.Process(ctx, recordBatchUnit{record: r, batchOperation: batchOperationDelete}); err != nil {
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

type batchOperation string

const (
	batchOperationCreate batchOperation = "create"
	batchOperationDelete batchOperation = "delete"
)

type recordBatchUnit struct {
	record         clouddnsv1.Record
	batchOperation batchOperation
}

func resourceDNSRecordBatch(a api.API, zoneName string) func(ctx context.Context, records []recordBatchUnit) []utils.BatchUnitResult[any] {
	return func(ctx context.Context, records []recordBatchUnit) []utils.BatchUnitResult[any] {
		res := make([]utils.BatchUnitResult[any], len(records))

		// ridiculously high timeout -> will be canceld before by schema timeout
		err := retry.RetryContext(ctx, time.Hour, func() *retry.RetryError {
			zone := clouddnsv1.Zone{Name: zoneName}
			if err := a.Get(ctx, &zone); err != nil {
				return retry.NonRetryableError(err)
			}

			if !zone.IsEditable {
				return retry.RetryableError(fmt.Errorf("zone not yet editable"))
			}

			return nil
		})
		if err != nil {
			for i := range records {
				res[i].Error = err
			}
		}

		changeSet := dnsZoneChangeSet{ZoneName: zoneName}
		for _, r := range records {
			changeSetRecord := dnsZoneChangeSetRecord{
				Name:   r.record.Name,
				Type:   r.record.Type,
				Region: r.record.Region,
				RData:  r.record.RData,
				TTL:    r.record.TTL,
			}
			if r.batchOperation == batchOperationCreate {
				changeSet.Create = append(changeSet.Create, changeSetRecord)
			} else if r.batchOperation == batchOperationDelete {
				changeSet.Delete = append(changeSet.Delete, changeSetRecord)
			}
		}

		if err := a.Create(ctx, &changeSet); err != nil {
			if changeSet.Error != nil {
				var (
					createIndex = 0
					deleteIndex = 0
				)

				for i := range records {
					var opErr map[string][]string
					if changeSet.Error.Create != nil && records[i].batchOperation == batchOperationCreate {
						opErr = changeSet.Error.Create[createIndex]
						createIndex++
					} else if changeSet.Error.Delete != nil && records[i].batchOperation == batchOperationDelete {
						opErr = changeSet.Error.Delete[deleteIndex]
						deleteIndex++
					}

					if len(opErr) > 0 {
						var combined *multierror.Error
						for fieldName, errors := range opErr {
							combined = multierror.Append(combined, fmt.Errorf("[%s: %s]", fieldName, strings.Join(errors, " - ")))
						}
						res[i].Error = combined
					} else {
						res[i].Error = fmt.Errorf("failed to %s dns record as part of batch, because other records are invalid", records[i].batchOperation)
					}
				}
			} else {
				for i := range records {
					res[i].Error = err
				}
			}
		}

		return res
	}
}

type dnsZoneChangeSetRecord struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Region string `json:"region,omitempty"`
	RData  string `json:"rdata"`
	TTL    int    `json:"ttl"`
}

type dnsZoneChangeSetError struct {
	Create []map[string][]string
	Delete []map[string][]string
}

// todo: move to go-anxcloud at a later time
type dnsZoneChangeSet struct {
	ZoneName string                   `json:"-"`
	Create   []dnsZoneChangeSetRecord `json:"create"`
	Delete   []dnsZoneChangeSetRecord `json:"delete"`
	Error    *dnsZoneChangeSetError   `json:"error,omitempty"`
}

func (cs *dnsZoneChangeSet) GetIdentifier(ctx context.Context) (string, error) {
	return "<not-used>", nil
}

func (cs *dnsZoneChangeSet) EndpointURL(ctx context.Context) (*url.URL, error) {
	op, err := types.OperationFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if op != types.OperationCreate {
		return nil, errors.New("helper resource 'dnsZoneChangeSet' only supports Create operations")
	}

	return url.Parse(fmt.Sprintf("/api/clouddns/v1/zone.json/%s/changeset", cs.ZoneName))
}

// FilterAPIResponse decodes record errors into the dnsZoneChangeSet so that we can output
// detailed error messages per zone instead of the same generic error message
func (cs *dnsZoneChangeSet) FilterAPIResponse(ctx context.Context, res *http.Response) (*http.Response, error) {
	if res.StatusCode == http.StatusOK {
		res.StatusCode = http.StatusNoContent
		res.Body.Close()
		res.Body = io.NopCloser(&bytes.Buffer{})
	} else if res.StatusCode == http.StatusBadRequest {
		if err := json.NewDecoder(res.Body).Decode(cs); err != nil {
			return nil, fmt.Errorf("unable to decode bad request response: %w", err)
		}
	}

	return res, nil
}

func resourceDNSRecordCanonicalIdentifier(r clouddnsv1.Record) string {
	return strings.Join([]string{
		r.Name,
		r.ZoneName,
		r.Type,
		url.QueryEscape(r.RData),
		fmt.Sprint(r.TTL),
		r.Region,
		fmt.Sprint(r.Immutable),
	}, "_")
}
