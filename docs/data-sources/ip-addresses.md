---
page_title: "ip_addresses Data Source - terraform-provider-anxcloud"
subcategory: ""
description: |-
  The IP addresses data source allows you to get all available IP addresses.
---

# Data Source `anxcloud_ip_addresses`

The IP addresses data source allows you to get all available IP addresses.

## Example Usage

```hcl
data "anxcloud_ip_addresses" "example" {
  search = "10.244."
}
```

## Argument Reference

- `page` - (Optional) The number of page. Defaults to 1.
- `limit` - (Optional) The records limit. Defaults to 1000.
- `search` - (Optional) The string allowing to search trough entities.

## Attributes Reference

The following attributes are exported.

- `addresses` - List of ip addresses. See [Addresses](#addresses) below for details.

### Addresses

- `identifier` - The identifier of the address.
- `address` - The IP address.
- `description_customer` - Additional customer description.
- `role` - Role of the IP address.
