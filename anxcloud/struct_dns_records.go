package anxcloud

import (
	clouddnsv1 "go.anx.io/go-anxcloud/pkg/apis/clouddns/v1"
	"go.anx.io/go-anxcloud/pkg/clouddns/zone"
)

func flattenDNSRecords(records []zone.Record) []interface{} {
	zoneRecords := make([]interface{}, 0, len(records))
	if len(records) < 1 {
		return zoneRecords
	}

	for _, record := range records {
		m := map[string]interface{}{
			"identifier": record.Identifier.String(),
			"immutable":  record.Immutable,
			"name":       record.Name,
			"rdata":      record.RData,
			"region":     record.Region,
			"type":       record.Type,
		}

		if record.TTL != nil {
			m["ttl"] = *record.TTL
		}

		zoneRecords = append(zoneRecords, m)
	}
	return zoneRecords
}

func flattenDNSRecordsV1(records []clouddnsv1.Record) []interface{} {
	zoneRecords := make([]interface{}, 0, len(records))
	if len(records) < 1 {
		return zoneRecords
	}

	for _, record := range records {
		m := map[string]interface{}{
			"identifier": record.Identifier,
			"immutable":  record.Immutable,
			"name":       record.Name,
			"rdata":      record.RData,
			"region":     record.Region,
			"type":       record.Type,
			"ttl":        record.TTL,
		}

		zoneRecords = append(zoneRecords, m)
	}
	return zoneRecords
}
