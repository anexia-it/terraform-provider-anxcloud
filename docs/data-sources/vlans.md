---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "anxcloud_vlans Data Source - terraform-provider-anxcloud"
subcategory: ""
description: |-
  Provides available VLANs.
---

# anxcloud_vlans (Data Source)

Provides available VLANs.

## Example Usage

```terraform
data "anxcloud_vlans" "example" {
  search = "tests"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `search` (String) An optional string allowing to search through entities.

### Read-Only

- `id` (String) The ID of this resource.
- `vlans` (List of Object) List of available VLANs. (see [below for nested schema](#nestedatt--vlans))

<a id="nestedatt--vlans"></a>
### Nested Schema for `vlans`

Read-Only:

- `description_customer` (String) Additional customer description.
- `identifier` (String) Identifier of the API resource.
- `name` (String) VLAN name.


