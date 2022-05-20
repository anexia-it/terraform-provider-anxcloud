---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "anxcloud_ip_address Data Source - terraform-provider-anxcloud"
subcategory: ""
description: |-
  Retrieves an IP address.
  Known limitations
  When using the address argument, only IP addresses unique to the scope of your access token for Anexia Cloud can be retrieved. You can however get a unique result by specifying the related VLAN or network prefix.
---

# anxcloud_ip_address (Data Source)

Retrieves an IP address.

### Known limitations

- When using the address argument, only IP addresses unique to the scope of your access token for Anexia Cloud can be retrieved. You can however get a unique result by specifying the related VLAN or network prefix.

## Example Usage

```terraform
data "anxcloud_ip_address" "example" {
  address = "10.244.2.50"
}

// This data-source can be used to create virtual servers.
// The following example is incomplete and won't work as is!
// Check out the anxcloud_virtual_server documentation for a full example.
resource "anxcloud_virtual_server" "example" {

  network {
    vlan_id  = data.anxcloud_ip_address.example.vlan_id
    ips      = [data.anxcloud_ip_address.example.address]
    nic_type = "vmxnet3"
  }

}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `address` (String) IP address.
- `id` (String) Identifier of the API resource.
- `network_prefix_id` (String) Identifier of the related network prefix.
- `vlan_id` (String) The associated VLAN identifier.

### Read-Only

- `description_customer` (String) Additional customer description.
- `description_internal` (String) Internal description.
- `organization` (String) Customer of yours. Reseller only.
- `role` (String) Role of the IP address
- `status` (String) Status of the IP address
- `version` (Number) IP version.

