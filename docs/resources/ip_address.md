---
page_title: "ip_address resource - terraform-provider-anxcloud"
subcategory: ""
description: |-
  The IP address resource allows you to create IP address at Anexia Cloud.
---

# Resource `anxcloud_ip_address`

-> Visit the [Perform CRUD operations with Providers](https://learn.hashicorp.com/tutorials/terraform/provider-use?in=terraform/providers&utm_source=WEBSITE&utm_medium=WEB_IO&utm_offer=ARTICLE_PAGE&utm_content=DOCS) Learn tutorial for an interactive getting started experience.

The IP address resource allows you to configure and create IP address at Anexia Cloud.

## Example Usage

```hcl
resource "anxcloud_ip_address" "example" {
  network_prefix_id    = "f4c2d6d415a0455d8ceb6bde09e4123e"
  address              = "185.160.55.162"
  description_customer = "example IP address"
}
```

## Argument Reference

- `network_prefix_id` - (Required) Network prefix identifier.
- `address` - (Required) The IP address that should be created.
- `role` - (Optional) Role of the IP address. Defaults to `Default`.
- `description_customer` - (Optional) Additional customer description.
- `organization` - (Optional) Customer of yours. Reseller only.

## Attributes Reference

In addition to all the arguments above, the following attributes are exported:

- `id` - IP address identifier.
- `description_internal` - Internal description of the network prefix.
- `status` - IP address status.
- `version` - IP address verion, either IPv6 or IPv4.
- `vlan_id` - The associated VLAN identifier.
