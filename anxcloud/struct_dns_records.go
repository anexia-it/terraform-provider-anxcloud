package anxcloud

import "github.com/anexia-it/go-anxcloud/pkg/clouddns/zone"

func flattenDnsRecords(records []zone.Record) []interface{} {
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
