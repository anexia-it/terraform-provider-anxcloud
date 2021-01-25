---
page_title: "anxcloud_tags Data Source - terraform-provider-anxcloud"
subcategory: ""
description: |-
  The tags data source allows you to get all available tags.
---

# Data Source `anxcloud_vsphere_locations`

The tags data source allows you to get all available tags.

## Example Usage

```hcl
data "anxcloud_tags" "example" {
  page  = 1
  limit = 100
}
```

## Argument Reference

- `page` - (Optional) The number of page. Defaults to 1.
- `limit` - (Optional) The records limit. Defaults to 1000.
- `query` - (Optional) Filters tages via search term.
- `service_identifier` - (Optional) Filters tags via service identifier.
- `organization_identifier` - (Optional) Filters tags via organization identifier.
- `order` - (Optional) Defines the order of the tags.
- `sort_ascending` - (Optional) Determines if the order of the tags is ascending or descending.

## Attributes Reference

The following attributes are exported.

- `tags` - List of tags. See [Tags](#tags) below for details.

### Locations

- `name` - Tag name.
- `identifier` - Tag identifier.
