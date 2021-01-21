---
page_title: "core_locations Data Source - terraform-provider-anxcloud"
subcategory: ""
description: |-
  The core locations data source allows you to get all available core locations.
---

# Data Source `anxcloud_core_locations`

The core locations data source allows you to get all available core locations.

## Example Usage

```hcl
data "anxcloud_core_locations" "example" {
  search = "IE"
}
```

## Argument Reference

- `page` - (Optional) The number of page. Defaults to 1.
- `limit` - (Optional) The records limit. Defaults to 1000.
- `search` - (Optional) The string allowing to search trough entities.

## Attributes Reference

The following attributes are exported.

- `locations` - List of locations. See [Locations](#locations) below for details.

### Locations

- `identifier` - Location identifier.
- `name` - Location name.
- `code` - Location code.
- `country` - Location country.
- `city_code` - Location city code.
- `lat` - Location latitude.
- `lon` - Location longitude.
