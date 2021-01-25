---
page_title: "anxcloud_cpu_performance_types Data Source - terraform-provider-anxcloud"
subcategory: ""
description: |-
  The cpu performance types data source allows you to get all available cpu performance types.
---

# Data Source `anxcloud_vsphere_locations`

The cpu performance types data source allows you to get all available cpu performance types.

## Example Usage

```hcl
data "anxcloud_cpu_performance_types" "example" {
  }
```

## Argument Reference

There are no arguments.

## Attributes Reference

The following attributes are exported.

- `types` - List of cpu performance types. See [CPU Performance Types](#Types) below for details.

### Types

- `id` - Id of the CPU performance type.
- `prioritization` - Prio of the CPU performance type.
- `limit` - The limit of the CPU performance type.
- `unit` - The unit for the limit of the CPU performance type.
