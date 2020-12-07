package anxcloud

import (
	"testing"

	"github.com/anexia-it/go-anxcloud/pkg/vsphere/provisioning/templates"
	"github.com/google/go-cmp/cmp"
)

// expanders tests

// flatteners tests

func TestFlattenTemplates(t *testing.T) {
	cases := []struct {
		Input          []templates.Template
		ExpectedOutput []interface{}
	}{
		{
			[]templates.Template{
				{
					ID:       "1233",
					Name:     "Centos 7",
					WordSize: "64",
					Build:    "b13",
				},
				{
					ID:       "1234",
					Name:     "Centos 7",
					WordSize: "32",
					Build:    "b14",
				},
			},
			[]interface{}{
				map[string]interface{}{
					"id":    "1233",
					"name":  "Centos 7",
					"bit":   "64",
					"build": "b13",
				},
				map[string]interface{}{
					"id":    "1234",
					"name":  "Centos 7",
					"bit":   "32",
					"build": "b14",
				},
			},
		},
		{
			[]templates.Template{},
			[]interface{}{},
		},
	}

	for _, tc := range cases {
		output := flattenTemplates(tc.Input)
		if diff := cmp.Diff(tc.ExpectedOutput, output); diff != "" {
			t.Fatalf("Unexpected output from expander: mismatch (-want +got):\n%s", diff)
		}
	}
}
