package anxcloud

import (
	"testing"

	"github.com/anexia-it/go-anxcloud/pkg/vsphere/provisioning/location"
	"github.com/google/go-cmp/cmp"
)

// expanders tests

// flatteners tests

func TestFlattenLocations(t *testing.T) {
	cases := []struct {
		Input          []location.Location
		ExpectedOutput []interface{}
	}{
		{
			[]location.Location{
				{
					Code:        "ADZ001",
					Country:     "AT",
					ID:          "5d80bbd5c69546218f7cb032d97fd067",
					Latitude:    "46.6364598000000",
					Longitude:   "14.3122246000000",
					Name:        "Anexia Deployment Zone 001 (ANX04/ANX88)",
					CountryName: "Austria",
				},
				{
					Code:        "ANX001",
					Country:     "AT",
					ID:          "5d80bbd5c69546218f7cb032d97fd068",
					Latitude:    "42.6364598000000",
					Longitude:   "15.3122246000000",
					Name:        "AT, Klagenfurt, STW",
					CountryName: "Austria",
				},
			},
			[]interface{}{
				map[string]interface{}{
					"code":         "ADZ001",
					"country":      "AT",
					"identifier":   "5d80bbd5c69546218f7cb032d97fd067",
					"lat":          "46.6364598000000",
					"lon":          "14.3122246000000",
					"name":         "Anexia Deployment Zone 001 (ANX04/ANX88)",
					"country_name": "Austria",
				},
				map[string]interface{}{
					"code":         "ANX001",
					"country":      "AT",
					"identifier":   "5d80bbd5c69546218f7cb032d97fd068",
					"lat":          "42.6364598000000",
					"lon":          "15.3122246000000",
					"name":         "AT, Klagenfurt, STW",
					"country_name": "Austria"},
			},
		},
		{
			[]location.Location{},
			[]interface{}{},
		},
	}

	for _, tc := range cases {
		output := flattenVSphereLocations(tc.Input)
		if diff := cmp.Diff(tc.ExpectedOutput, output); diff != "" {
			t.Fatalf("Unexpected output from expander: mismatch (-want +got):\n%s", diff)
		}
	}
}
