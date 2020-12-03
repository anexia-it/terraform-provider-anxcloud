---
page_title: "disk_type Data Source - terraform-provider-anxcloud"
subcategory: ""
description: |-
  The disk type data source allows you to retrieve information about available disk types for specified location.
---

# Data Source `disk_type`

The disk type data source allows you to retrieve information about available disk types for specified location.

## Example Usage

```hcl
data "anxcloud_disk_type" "example" {
  location_id = "52b5f6b2fd3a4a7eaaedf1a7c019e9ea"
}
```

## Argument Reference

- `location_id` - (Required) Location identifier.

## Attributes Reference

The following attributes are exported.

- `types` - A list of disk types objects. See [Types](#types) below for details.

### Types

- `id` -  Identifier of the disk type.
- `storage_type` - The disk storage type.
- `bandwidth` - The disk bandwidth.
- `iops` - Disk input/output operations per second.
- `latency` - The disk latency.
