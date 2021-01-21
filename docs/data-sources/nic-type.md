---
page_title: "NIC type Data Source - terraform-provider-anxcloud"
subcategory: ""
description: |-
  The NIC type data source allows you to get all available network interface card types.
---

# Data Source `anxcloud_nic_type`

The NIC type data source allows you to get all available network interface card types.

## Example Usage

```hcl
data "anxcloud_nic_type" "example" {}
```

## Attributes Reference

The following attributes are exported.

- `nic_types` - List of nic types.
