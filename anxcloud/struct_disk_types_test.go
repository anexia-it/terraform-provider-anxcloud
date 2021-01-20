package anxcloud

import (
	"testing"

	"github.com/anexia-it/go-anxcloud/pkg/vsphere/provisioning/disktype"
	"github.com/google/go-cmp/cmp"
)

// expanders tests

// flatteners tests

func TestFlattenDiskTypes(t *testing.T) {
	cases := []struct {
		Input          []disktype.DiskType
		ExpectedOutput []interface{}
	}{
		{
			[]disktype.DiskType{
				{
					Bandwidth:   300,
					ID:          "STD6",
					IOPS:        2600,
					Latency:     30,
					StorageType: "HDD",
				},
				{
					Bandwidth:   500,
					ID:          "HPC1",
					IOPS:        20000,
					Latency:     7,
					StorageType: "SSD",
				},
			},
			[]interface{}{
				map[string]interface{}{
					"bandwidth":    300,
					"id":           "STD6",
					"iops":         2600,
					"latency":      30,
					"storage_type": "HDD",
				},
				map[string]interface{}{
					"bandwidth":    500,
					"id":           "HPC1",
					"iops":         20000,
					"latency":      7,
					"storage_type": "SSD",
				},
			},
		},
		{
			[]disktype.DiskType{},
			[]interface{}{},
		},
	}

	for _, tc := range cases {
		output := flattenDiskTypes(tc.Input)
		if diff := cmp.Diff(tc.ExpectedOutput, output); diff != "" {
			t.Fatalf("Unexpected output from expander: mismatch (-want +got):\n%s", diff)
		}
	}
}
