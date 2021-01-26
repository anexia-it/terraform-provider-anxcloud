---
page_title: "vsphere_locations Data Source - terraform-provider-anxcloud"
subcategory: ""
description: |-
  The vsphere locations data source allows you to get all available vsphere locations.
---

# Data Source `anxcloud_vsphere_locations`

The vsphere locations data source allows you to get all available vsphere locations.

## Example Usage

```hcl
data "anxcloud_vsphere_locations" "example" {
  page  = 1
  limit = 50
}
```

## Argument Reference

- `page` - (Optional) The number of page. Defaults to 1.
- `limit` - (Optional) The records limit. Defaults to 50.
- `location_code` - (Optional) Filters locations by country code.
- `organization` - (Optional) Filters locations by customer identifier.

## Attributes Reference

The following attributes are exported.

- `locations` - List of locations. See [Locations](#locations) below for details.

### Locations

- `identifier` - Location identifier.
- `name` - Location name.
- `code` - Location code.
- `country` - Location country.
- `country_name` - Location country name.
- `lat` - Location latitude.
- `lon` - Location longitude.
