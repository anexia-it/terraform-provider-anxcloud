---
page_title: "network_prefix resource - terraform-provider-anxcloud"
subcategory: ""
description: |-
  The network prefix resource allows you to create network prefix at Anexia Cloud.
---

# Resource `anxcloud_network_prefix`

-> Visit the [Perform CRUD operations with Providers](https://learn.hashicorp.com/tutorials/terraform/provider-use?in=terraform/providers&utm_source=WEBSITE&utm_medium=WEB_IO&utm_offer=ARTICLE_PAGE&utm_content=DOCS) Learn tutorial for an interactive getting started experience.

The network prefix resource allows you to configure and create network prefix at Anexia Cloud.

## Example Usage

```hcl
resource "anxcloud_network_prefix" "example" {
  vlan_id     = "e3c2d6d415a0455d8ceb6bde09e4d08e"
  location_id = "52b5f6b2fd3a4a7eaaedf1a7c019e9ea"
  ip_version  = 4
  type        = 0
  netmask     = 29
}
```

## Argument Reference

- `location_id` - (Required) Location identifier.
- `netmask` - (Required) Netmask size. Example: 29 which would result in x.x.x.x/29.
- `ip_version` - (Optional) The Prefix version: 4 = IPv4, 6 = IPv6.
- `type` - (Optional) The Prefix type: 0 = Public, 1 = Private.
- `vlan_id` - (Optional) Identifier for the related VLAN. Not applicable when using `new_vlan` option.
- `router_redundancy` - (Optional) If router Redundancy shall be enabled.
- `description_customer` - (Optional) Additional customer description.
- `organization` - (Optional) Customer of yours. Reseller only.


## Attributes Reference

In addition to all the arguments above, the following attributes are exported:

- `id` - Network prefix identifier.
- `cidr` - CIDR of the created network prefix.
- `description_internal` - Internal description of the network prefix.
- `role_text` - Role of the network prefix in text format.
- `status` - Network prefix status.
- `locations` - Network prefix locations. See [locations](#locations) below for details.

### Locations

- `identifier` - Location identifier.
- `name` - Location name.
- `code` - Location code.
- `country` - Location country.
- `city_code` - Location city code.
- `lat` - Location latitude.
- `lon` - Location longitude.
