package anxcloud

import (
	"github.com/anexia-it/go-anxcloud/pkg/ipam/prefix"
)

// expanders

// flatteners

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
