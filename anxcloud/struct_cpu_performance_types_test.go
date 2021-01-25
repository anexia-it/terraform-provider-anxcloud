package anxcloud

import (
	"testing"

	cpuperformancetype "github.com/anexia-it/go-anxcloud/pkg/vsphere/provisioning/cpuperformancetypes"
	"github.com/google/go-cmp/cmp"
)

// expanders tests

// flatteners tests

func TestFlattenCPUPerfrormanceTypes(t *testing.T) {
	cases := []struct {
		Input          []cpuperformancetype.CPUPerformanceType
		ExpectedOutput []interface{}
	}{
		{
			[]cpuperformancetype.CPUPerformanceType{
				{
					ID:             "best-effort",
					Prioritization: "Low",
					Limit:          0.5,
					Unit:           "GHz",
				},
				{
					ID:             "standard",
					Prioritization: "Medium",
					Limit:          1.8,
					Unit:           "GHz",
				},
			},
			[]interface{}{
				map[string]interface{}{
					"id":             "best-effort",
					"prioritization": "Low",
					"limit":          0.5,
					"unit":           "GHz",
				},
				map[string]interface{}{
					"id":             "standard",
					"prioritization": "Medium",
					"limit":          1.8,
					"unit":           "GHz",
				},
			},
		},
		{
			[]cpuperformancetype.CPUPerformanceType{},
			[]interface{}{},
		},
	}

	for _, tc := range cases {
		output := flattenCPUPerformanceTypes(tc.Input)
		if diff := cmp.Diff(tc.ExpectedOutput, output); diff != "" {
			t.Fatalf("Unexpected output from expander: mismatch (-want +got):\n%s", diff)
		}
	}
}
