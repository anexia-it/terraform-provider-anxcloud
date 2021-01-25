package anxcloud

import (
	corelocation "github.com/anexia-it/go-anxcloud/pkg/core/location"
	"github.com/anexia-it/go-anxcloud/pkg/ipam/prefix"
	"github.com/anexia-it/go-anxcloud/pkg/vlan"
	"github.com/anexia-it/go-anxcloud/pkg/vsphere/provisioning/location"
)

// expanders

// flatteners

// TODO: we have a few structures in go-sdk for the same locations, this must be fixed there and later here

func flattenCoreLocations(in []corelocation.Location) []interface{} {
	att := []interface{}{}
	if len(in) < 1 {
		return att
	}

	for _, l := range in {
		m := map[string]interface{}{}

		m["identifier"] = l.ID
		m["name"] = l.Name
		m["code"] = l.Code
		m["city_code"] = l.CityCode
		m["country"] = l.Country
		m["lat"] = l.Latitude
		m["lon"] = l.Longitude

		att = append(att, m)
	}

	return att
}

func flattenNetworkPrefixLocations(in []prefix.Location) []interface{} {
	att := []interface{}{}
	if len(in) < 1 {
		return att
	}

	for _, l := range in {
		m := map[string]interface{}{}

		m["identifier"] = l.ID
		m["name"] = l.Name
		m["code"] = l.Code
		m["city_code"] = l.CityCode
		m["country"] = l.Country
		m["lat"] = l.Latitude
		m["lon"] = l.Longitude

		att = append(att, m)
	}

	return att
}

func flattenVLANLocations(in []vlan.Location) []interface{} {
	att := []interface{}{}
	if len(in) < 1 {
		return att
	}

	for _, l := range in {
		m := map[string]interface{}{}

		m["identifier"] = l.Identifier
		m["name"] = l.Name
		m["code"] = l.Code
		m["city_code"] = l.CityCode
		m["country"] = l.CountryCode
		m["lat"] = l.Latitude
		m["lon"] = l.Longitude

		att = append(att, m)
	}

	return att
}

func flattenVSphereLocations(in []location.Location) []interface{} {

	att := []interface{}{}
	if len(in) < 1 {
		return att
	}

	for _, d := range in {
		m := map[string]interface{}{}

		m["code"] = d.Code
		m["country"] = d.Country
		m["identifier"] = d.ID
		m["lat"] = d.Latitude
		m["lon"] = d.Longitude
		m["name"] = d.Name
		m["country_name"] = d.CountryName

		att = append(att, m)
	}

	return att

}
