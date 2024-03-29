---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "anxcloud_ip_address Resource - terraform-provider-anxcloud"
subcategory: ""
description: |-
  This resource allows you to create and configure IP addresses.
---

# anxcloud_ip_address (Resource)

This resource allows you to create and configure IP addresses.

## Example Usage

```terraform
data "anxcloud_core_location" "anx04" {
  code = "ANX04"
}

resource "anxcloud_ip_address" "example" {
  network_prefix_id = var.network_prefix_id
  address           = "10.20.30.1"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `address` (String) IP address.
- `network_prefix_id` (String) Identifier of the related network prefix.

### Optional

- `description_customer` (String) Additional customer description.
- `organization` (String) Customer of yours. Reseller only.
- `role` (String) Role of the IP address
- `tags` (Set of String) Set of tags attached to the resource.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `description_internal` (String) Internal description.
- `id` (String) Identifier of the API resource.
- `status` (String) Status of the IP address
- `version` (Number) IP version.
- `vlan_id` (String) The associated VLAN identifier.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)
- `delete` (String)
- `read` (String)
- `update` (String)


