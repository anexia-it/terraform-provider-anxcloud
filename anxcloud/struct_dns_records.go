package anxcloud

import "github.com/anexia-it/go-anxcloud/pkg/clouddns/zone"

func flattenDnsRecords(records []zone.Record) []interface{} {
	zoneRecords := []interface{}{}
	if len(records) < 1 {
		return zoneRecords
	}

	for _, record := range records {
		m := map[string]interface{}{}

		m["identifier"] = record.Identifier.String()
		m["immutable"] = record.Immutable
		m["name"] = record.Name
		m["rdata"] = record.RData
		m["region"] = record.Region
		if record.TTL != nil {
			m["ttl"] = record.TTL
		}
		m["type"] = record.Type

		zoneRecords = append(zoneRecords, m)
	}
	return zoneRecords
}
