package anxcloud

import (
	"github.com/anexia-it/go-anxcloud/pkg/vsphere/provisioning/location"
)

// expanders

// flatteners

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
