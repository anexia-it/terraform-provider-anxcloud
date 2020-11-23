---
page_title: "vlan resource - terraform-provider-anxcloud"
subcategory: ""
description: |-
  The VLAN resource allows you to create VLAN at Anexia Cloud.
---

# Resource `anxcloud_vlan`

-> Visit the [Perform CRUD operations with Providers](https://learn.hashicorp.com/tutorials/terraform/provider-use?in=terraform/providers&utm_source=WEBSITE&utm_medium=WEB_IO&utm_offer=ARTICLE_PAGE&utm_content=DOCS) Learn tutorial for an interactive getting started experience.

The VLAN resource allows you to configure and create VLAN at Anexia Cloud.

## Example Usage

```hcl
resource "anxcloud_vlan" "example" {
    location_id = "52b5f6b2fd3a4a7eaaedf1a7c019e9ea"
    vm_provisioning = true
    description_customer = "sample vlan"
}
```

## Argument Reference

- `location_id` - (Required) Location identifier.
- `vm_provisioning` - (Optional) True if VM provisioning shall be enabled. Defaults to false.
- `description_customer` - (Optional) Additional customer description.

## Attributes Reference

In addition to all the arguments above, the following attributes are exported:

- `id` - VLAN identifier.
- `name` - Generated VLAN name.
- `description_internal` - Internal description of the VLAN.
- `role_text` - Role of the VLAN in text format.
- `status` - VLAN status.
- `locations` - VLAN locations. See [locations](#locations) below for details.

### Locations

- `identifier` - Location identifier.
- `name` - Location name.
- `code` - Location code.
- `country` - Location country.
- `city_code` - Location city code.
- `lat` - Location latitude.
- `lon` - Location longitude.
