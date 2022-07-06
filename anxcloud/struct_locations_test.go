package anxcloud

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	corelocation "go.anx.io/go-anxcloud/pkg/core/location"
	"go.anx.io/go-anxcloud/pkg/ipam/prefix"
	"go.anx.io/go-anxcloud/pkg/vlan"
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
