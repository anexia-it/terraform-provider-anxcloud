// Package anxcloud provides Terraform resources for Anexia Cloud services.
// This file contains DNS record matching and search utilities.
package anxcloud

import (
	"context"

	"go.anx.io/go-anxcloud/pkg/api"
	"go.anx.io/go-anxcloud/pkg/api/types"
	clouddnsv1 "go.anx.io/go-anxcloud/pkg/apis/clouddns/v1"
	"go.anx.io/go-anxcloud/pkg/utils/object/compare"
)

// fetchAllZoneRecords retrieves all DNS records for a given zone
func fetchAllZoneRecords(ctx context.Context, a api.API, zoneName string) ([]clouddnsv1.Record, error) {
	var pageIter types.PageInfo

	// Create a dummy record for zone context
	dummyRecord := clouddnsv1.Record{ZoneName: zoneName}

	err := a.List(ctx, &dummyRecord, api.Paged(1, 100, &pageIter))
	if err != nil {
		return nil, err
	}

	// Collect all records from all pages
	allRecords := make([]clouddnsv1.Record, 0, pageIter.TotalItems())
	var pagedRecords []clouddnsv1.Record
	for pageIter.Next(&pagedRecords) {
		allRecords = append(allRecords, pagedRecords...)
	}

	if err := pageIter.Error(); err != nil {
		return nil, err
	}

	return allRecords, nil
}

// findRecordByContent finds a record in the zone by matching content fields
func findRecordByContent(records []clouddnsv1.Record, target clouddnsv1.Record) (*clouddnsv1.Record, error) {
	return findRecordByContentFlexible(records, target, false, false)
}

// isQuotedString checks if a string starts and ends with double quotes
func isQuotedString(s string) bool {
	return len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"'
}

// stripOuterQuotes removes the outer quotes from a quoted string
// Returns the original string if it's not properly quoted
func stripOuterQuotes(s string) string {
	if isQuotedString(s) {
		return s[1 : len(s)-1]
	}
	return s
}

// findRecordByContentFlexible finds a record with optional flexible matching for TTL and region
func findRecordByContentFlexible(records []clouddnsv1.Record, target clouddnsv1.Record, ignoreTTL, ignoreRegion bool) (*clouddnsv1.Record, error) {
	// Compare records by content fields
	// RData is compared as-is (no quote manipulation needed)
	targetRData := target.RData

	for _, record := range records {
		recordRData := record.RData

		// Match by key fields with optional flexibility
		ttlMatches := ignoreTTL || record.TTL == target.TTL
		regionMatches := ignoreRegion || record.Region == target.Region

		if record.Type == target.Type &&
			record.Name == target.Name &&
			recordRData == targetRData &&
			ttlMatches &&
			regionMatches &&
			record.ZoneName == target.ZoneName {
			return &record, nil
		}
	}

	return nil, api.ErrNotFound
}

func findDNSRecord(ctx context.Context, a api.API, r clouddnsv1.Record) (foundRecord clouddnsv1.Record, err error) {
	// quote TXTs rdata for compare.Compare (SYSENG-816)
	// We manually add quotes instead of using %q to preserve user's quote characters
	if r.Type == "TXT" {
		r.RData = `"` + r.RData + `"`
	}

	// For the API List call, only use ZoneName to avoid filtering out records
	// We'll search by other fields manually after fetching all zone records
	listQuery := clouddnsv1.Record{ZoneName: r.ZoneName}

	var pageIter types.PageInfo
	err = a.List(ctx, &listQuery, api.Paged(1, 100, &pageIter))
	if err != nil {
		foundRecord = clouddnsv1.Record{}
		return foundRecord, err
	}

	var pagedRecords []clouddnsv1.Record
	for pageIter.Next(&pagedRecords) {
		// If we have an identifier, search by identifier first (most efficient)
		if r.Identifier != "" {
			for _, record := range pagedRecords {
				if record.Identifier == r.Identifier {
					foundRecord = record
					return foundRecord, nil
				}
			}
		}

		// Fall back to content-based search
		idx, err := compare.Search(&r, pagedRecords, "Type", "Name", "RData", "TTL")
		if err != nil {
			foundRecord = clouddnsv1.Record{}
			return foundRecord, err
		}
		if idx > -1 {
			foundRecord = pagedRecords[idx]
			return foundRecord, nil
		}
	}

	foundRecord = clouddnsv1.Record{}
	return foundRecord, api.ErrNotFound
}
