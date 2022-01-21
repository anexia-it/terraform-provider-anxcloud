package anxcloud

import (
	"testing"

	"go.anx.io/go-anxcloud/pkg/ipam/address"

	"github.com/google/go-cmp/cmp"
)

// expanders tests

// flatteners tests

func TestFlattenIPAddresses(t *testing.T) {
	cases := []struct {
		Input          []address.Summary
		ExpectedOutput []interface{}
	}{
		{
			[]address.Summary{
				{
					ID:                  "d866b766d71947e3aac7f60409383b42",
					Name:                "10.244.2.24",
					DescriptionCustomer: "project x",
					Role:                "Default",
				},
				{
					ID:                  "e866b766d71947e3aac7f60409383b62",
					Name:                "10.244.2.25",
					DescriptionCustomer: "",
					Role:                "Network",
				},
			},
			[]interface{}{
				map[string]interface{}{
					"identifier":           "d866b766d71947e3aac7f60409383b42",
					"address":              "10.244.2.24",
					"description_customer": "project x",
					"role":                 "Default",
				},
				map[string]interface{}{
					"identifier":           "e866b766d71947e3aac7f60409383b62",
					"address":              "10.244.2.25",
					"description_customer": "",
					"role":                 "Network",
				},
			},
		},
		{
			[]address.Summary{},
			[]interface{}{},
		},
	}

	for _, tc := range cases {
		output := flattenIPAddresses(tc.Input)
		if diff := cmp.Diff(tc.ExpectedOutput, output); diff != "" {
			t.Fatalf("Unexpected output from expander: mismatch (-want +got):\n%s", diff)
		}
	}
}
