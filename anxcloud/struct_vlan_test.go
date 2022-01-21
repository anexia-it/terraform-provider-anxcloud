package anxcloud

import (
	"testing"

	"go.anx.io/go-anxcloud/pkg/vlan"

	"github.com/google/go-cmp/cmp"
)

// expanders tests

// flatteners tests

func TestFlattenVLANs(t *testing.T) {
	cases := []struct {
		Input          []vlan.Summary
		ExpectedOutput []interface{}
	}{
		{
			[]vlan.Summary{
				{
					Identifier:          "d866b766d71947e3aac7f60409383b45",
					Name:                "VLAN3091",
					CustomerDescription: "project x",
				},
				{
					Identifier:          "e866b766d71947e3aac7f60409383b68",
					Name:                "VLAN3092",
					CustomerDescription: "",
				},
			},
			[]interface{}{
				map[string]interface{}{
					"identifier":           "d866b766d71947e3aac7f60409383b45",
					"name":                 "VLAN3091",
					"description_customer": "project x",
				},
				map[string]interface{}{
					"identifier":           "e866b766d71947e3aac7f60409383b68",
					"name":                 "VLAN3092",
					"description_customer": "",
				},
			},
		},
		{
			[]vlan.Summary{},
			[]interface{}{},
		},
	}

	for _, tc := range cases {
		output := flattenVLANs(tc.Input)
		if diff := cmp.Diff(tc.ExpectedOutput, output); diff != "" {
			t.Fatalf("Unexpected output from expander: mismatch (-want +got):\n%s", diff)
		}
	}
}
