---
page_title: "VLANs Data Source - terraform-provider-anxcloud"
subcategory: ""
description: |-
  The VLANs data source allows you to get all available VLANs.
---

# Data Source `anxcloud_vlans`

The VLANs data source allows you to get all available VLANs.

## Example Usage

```hcl
data "anxcloud_vlans" "example" {
  search = "tests"
}
```

## Argument Reference

- `page` - (Optional) The number of page. Defaults to 1.
- `limit` - (Optional) The records limit. Defaults to 1000.
- `search` - (Optional) The string allowing to search trough entities.

## Attributes Reference

The following attributes are exported.

- `vlans` - List of VLANs. See [VLANs](#vlans) below for details.

### VLANs

- `identifier` - VLAN identifier.
- `name` - Generated VLAN name.
- `description_customer` - Additional customer description.
