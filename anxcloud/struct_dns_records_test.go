package anxcloud

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	uuid "github.com/satori/go.uuid"
	"go.anx.io/go-anxcloud/pkg/clouddns/zone"
)

func TestFlattenDnsRecords(t *testing.T) {
	id, _ := uuid.NewV4()
	id2, _ := uuid.NewV4()
	ttl := 100
	cases := []struct {
		Input          []zone.Record
		ExpectedOutput []interface{}
	}{
		{
			[]zone.Record{
				{
					Type:       "TXT",
					Name:       "test-record-1",
					RData:      "127.0.0.1",
					Region:     "DACH",
					Immutable:  false,
					Identifier: id,
					TTL:        &ttl,
				},
				{
					Type:       "TXT",
					Name:       "test-record-2",
					RData:      "127.0.0.2",
					Region:     "EU",
					Immutable:  false,
					Identifier: id2,
				},
			},
			[]interface{}{
				map[string]interface{}{
					"identifier": id.String(),
					"name":       "test-record-1",
					"type":       "TXT",
					"rdata":      "127.0.0.1",
					"region":     "DACH",
					"immutable":  false,
					"ttl":        ttl,
				},
				map[string]interface{}{
					"identifier": id2.String(),
					"name":       "test-record-2",
					"type":       "TXT",
					"rdata":      "127.0.0.2",
					"region":     "EU",
					"immutable":  false,
				},
			},
		},
		{
			[]zone.Record{},
			[]interface{}{},
		},
	}

	for _, tc := range cases {
		output := flattenDNSRecords(tc.Input)
		if diff := cmp.Diff(tc.ExpectedOutput, output); diff != "" {
			t.Fatalf("Unexpected output from flattener: missmatch (-want +got):\n%s", diff)
		}
	}
}
