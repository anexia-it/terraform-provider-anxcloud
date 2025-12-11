// Package anxcloud provides Terraform resources for Anexia Cloud services.
// This file contains batch processing logic for DNS record operations.
package anxcloud

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/anexia-it/terraform-provider-anxcloud/anxcloud/internal/utils"
	"github.com/cenkalti/backoff/v4"
	"github.com/hashicorp/go-multierror"
	"go.anx.io/go-anxcloud/pkg/api"
	"go.anx.io/go-anxcloud/pkg/api/types"
	clouddnsv1 "go.anx.io/go-anxcloud/pkg/apis/clouddns/v1"
)

func resourceDNSRecordBatcherForZone(a api.API, zoneName string) *utils.Batcher[recordBatchUnit, any] {
	anyBatcher, _ := resourceDNSRecordBatcherMap.LoadOrStore(zoneName, &utils.Batcher[recordBatchUnit, any]{
		// this will consume 15 seconds of the 2 minute create/delete budget
		Wait:      15 * time.Second,
		BatchFunc: resourceDNSRecordBatch(a, zoneName),
	})

	return anyBatcher.(*utils.Batcher[recordBatchUnit, any])
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
		log.Printf("[DEBUG] DNS Record Batch: processing %d records for zone %s", len(records), zoneName)
		res := make([]utils.BatchUnitResult[any], len(records))

		// Wait for zone to be editable and fully deployed
		b := backoff.NewExponentialBackOff()
		b.InitialInterval = 10 * time.Second
		b.MaxInterval = 30 * time.Second
		b.MaxElapsedTime = 10 * time.Minute

		err := backoff.Retry(func() error {
			zone := clouddnsv1.Zone{Name: zoneName}
			if err := a.Get(ctx, &zone); err != nil {
				return backoff.Permanent(err)
			}

			if !zone.IsEditable {
				return fmt.Errorf("zone not yet editable")
			}

			return nil
		}, backoff.WithContext(b, ctx))
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
				log.Printf("[DEBUG] DNS Record Batch: adding create operation for %s %s %s", r.record.Name, r.record.Type, r.record.RData)
			} else if r.batchOperation == batchOperationDelete {
				changeSet.Delete = append(changeSet.Delete, changeSetRecord)
				log.Printf("[DEBUG] DNS Record Batch: adding delete operation for %s %s %s", r.record.Name, r.record.Type, r.record.RData)
			}
		}

		log.Printf("[DEBUG] DNS Record Batch: executing changeset with %d creates and %d deletes", len(changeSet.Create), len(changeSet.Delete))

		// Get zone state before changeset execution to verify it changes
		preChangeSetZone := clouddnsv1.Zone{Name: zoneName}
		if err := a.Get(ctx, &preChangeSetZone); err != nil {
			for i := range records {
				res[i].Error = fmt.Errorf("failed to get pre-changeset zone state: %w", err)
			}
			return res
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
		} else {
			// Changeset executed successfully - verify zone state changed
			log.Printf("[DEBUG] DNS Record Batch: changeset executed successfully, verifying zone state change")

			// Brief delay to allow zone state to update
			time.Sleep(2 * time.Second)

			postChangeSetZone := clouddnsv1.Zone{Name: zoneName}
			if err := a.Get(ctx, &postChangeSetZone); err != nil {
				for i := range records {
					res[i].Error = fmt.Errorf("failed to verify post-changeset zone state: %w", err)
				}
				return res
			}

			// Verify that changeset actually triggered zone validation or state change
			if postChangeSetZone.ValidationLevel == 0 && postChangeSetZone.DeploymentLevel == 100 &&
				preChangeSetZone.ValidationLevel == postChangeSetZone.ValidationLevel &&
				preChangeSetZone.DeploymentLevel == postChangeSetZone.DeploymentLevel {
				// Zone state unchanged - changeset may not have triggered validation
				log.Printf("[WARN] DNS Record Batch: changeset executed but zone state unchanged (validation: %d%%, deployment: %d%%) - validation may not have started",
					postChangeSetZone.ValidationLevel, postChangeSetZone.DeploymentLevel)
				// Don't fail here - let the individual record waiting logic handle it
			} else {
				log.Printf("[DEBUG] DNS Record Batch: zone state changed after changeset (validation: %d%% -> %d%%, deployment: %d%% -> %d%%)",
					preChangeSetZone.ValidationLevel, postChangeSetZone.ValidationLevel,
					preChangeSetZone.DeploymentLevel, postChangeSetZone.DeploymentLevel)
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
