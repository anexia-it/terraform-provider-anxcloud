package anxcloud

import (
	"github.com/anexia-it/go-anxcloud/pkg/vlan"
)

// expanders

// flatteners

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
