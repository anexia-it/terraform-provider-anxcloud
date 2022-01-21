package anxcloud

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	corelocation "go.anx.io/go-anxcloud/pkg/core/location"
	"go.anx.io/go-anxcloud/pkg/ipam/prefix"
	"go.anx.io/go-anxcloud/pkg/vlan"
	"go.anx.io/go-anxcloud/pkg/vsphere/provisioning/location"
)

// expanders tests

// flatteners tests

func TestFlattenCoreLocations(t *testing.T) {
	cases := []struct {
		Input          []corelocation.Location
		ExpectedOutput []interface{}
	}{
		{
			[]corelocation.Location{
				{
					ID:        "52b5f6b2fd3a4a7eaaedf1a7c0191234",
					Name:      "AT, Vienna, Test",
					Code:      "ANX-8888",
					Country:   "AT",
					Latitude:  "0.0",
					Longitude: "0.0",
					CityCode:  "VIE",
				},
				{
					ID:        "72c5f6b2fd3a4a7eaaedf1a7c0194321",
					Name:      "AT, Vienna, Test2",
					Code:      "ANX-8889",
					Country:   "AT",
					Latitude:  "1.1",
					Longitude: "1.1",
					CityCode:  "VIE",
				},
			},
			[]interface{}{
				map[string]interface{}{
					"identifier": "52b5f6b2fd3a4a7eaaedf1a7c0191234",
					"name":       "AT, Vienna, Test",
					"code":       "ANX-8888",
					"country":    "AT",
					"lat":        "0.0",
					"lon":        "0.0",
					"city_code":  "VIE",
				},
				map[string]interface{}{
					"identifier": "72c5f6b2fd3a4a7eaaedf1a7c0194321",
					"name":       "AT, Vienna, Test2",
					"code":       "ANX-8889",
					"country":    "AT",
					"lat":        "1.1",
					"lon":        "1.1",
					"city_code":  "VIE",
				},
			},
		},
		{
			[]corelocation.Location{},
			[]interface{}{},
		},
	}

	for _, tc := range cases {
		output := flattenCoreLocations(tc.Input)
		if diff := cmp.Diff(tc.ExpectedOutput, output); diff != "" {
			t.Fatalf("Unexpected output from expander: mismatch (-want +got):\n%s", diff)
		}
	}
}

func TestFlattenNetworkPrefixLocations(t *testing.T) {
	cases := []struct {
		Input          []prefix.Location
		ExpectedOutput []interface{}
	}{
		{
			[]prefix.Location{
				{
					ID:        "52b5f6b2fd3a4a7eaaedf1a7c0191234",
					Name:      "AT, Vienna, Test",
					Code:      "ANX-8888",
					Country:   "AT",
					Latitude:  "0.0",
					Longitude: "0.0",
					CityCode:  "VIE",
				},
				{
					ID:        "72c5f6b2fd3a4a7eaaedf1a7c0194321",
					Name:      "AT, Vienna, Test2",
					Code:      "ANX-8889",
					Country:   "AT",
					Latitude:  "1.1",
					Longitude: "1.1",
					CityCode:  "VIE",
				},
			},
			[]interface{}{
				map[string]interface{}{
					"identifier": "52b5f6b2fd3a4a7eaaedf1a7c0191234",
					"name":       "AT, Vienna, Test",
					"code":       "ANX-8888",
					"country":    "AT",
					"lat":        "0.0",
					"lon":        "0.0",
					"city_code":  "VIE",
				},
				map[string]interface{}{
					"identifier": "72c5f6b2fd3a4a7eaaedf1a7c0194321",
					"name":       "AT, Vienna, Test2",
					"code":       "ANX-8889",
					"country":    "AT",
					"lat":        "1.1",
					"lon":        "1.1",
					"city_code":  "VIE",
				},
			},
		},
		{
			[]prefix.Location{},
			[]interface{}{},
		},
	}

	for _, tc := range cases {
		output := flattenNetworkPrefixLocations(tc.Input)
		if diff := cmp.Diff(tc.ExpectedOutput, output); diff != "" {
			t.Fatalf("Unexpected output from expander: mismatch (-want +got):\n%s", diff)
		}
	}
}

func TestFlattenVLANLocations(t *testing.T) {
	cases := []struct {
		Input          []vlan.Location
		ExpectedOutput []interface{}
	}{
		{
			[]vlan.Location{
				{
					Identifier:  "52b5f6b2fd3a4a7eaaedf1a7c0191234",
					Name:        "AT, Vienna, Test",
					Code:        "ANX-8888",
					CountryCode: "AT",
					Latitude:    "0.0",
					Longitude:   "0.0",
					CityCode:    "VIE",
				},
				{
					Identifier:  "72c5f6b2fd3a4a7eaaedf1a7c0194321",
					Name:        "AT, Vienna, Test2",
					Code:        "ANX-8889",
					CountryCode: "AT",
					Latitude:    "1.1",
					Longitude:   "1.1",
					CityCode:    "VIE",
				},
			},
			[]interface{}{
				map[string]interface{}{
					"identifier": "52b5f6b2fd3a4a7eaaedf1a7c0191234",
					"name":       "AT, Vienna, Test",
					"code":       "ANX-8888",
					"country":    "AT",
					"lat":        "0.0",
					"lon":        "0.0",
					"city_code":  "VIE",
				},
				map[string]interface{}{
					"identifier": "72c5f6b2fd3a4a7eaaedf1a7c0194321",
					"name":       "AT, Vienna, Test2",
					"code":       "ANX-8889",
					"country":    "AT",
					"lat":        "1.1",
					"lon":        "1.1",
					"city_code":  "VIE",
				},
			},
		},
		{
			[]vlan.Location{},
			[]interface{}{},
		},
	}

	for _, tc := range cases {
		output := flattenVLANLocations(tc.Input)
		if diff := cmp.Diff(tc.ExpectedOutput, output); diff != "" {
			t.Fatalf("Unexpected output from expander: mismatch (-want +got):\n%s", diff)
		}
	}
}

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
