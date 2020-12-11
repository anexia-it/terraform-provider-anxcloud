---
page_title: "tag resource - terraform-provider-anxcloud"
subcategory: ""
description: |-
  The tag resource allows you to create a tag at Anexia Cloud.
---

# Resource `anxcloud_tag`

-> Visit the [Perform CRUD operations with Providers](https://learn.hashicorp.com/tutorials/terraform/provider-use?in=terraform/providers&utm_source=WEBSITE&utm_medium=WEB_IO&utm_offer=ARTICLE_PAGE&utm_content=DOCS) Learn tutorial for an interactive getting started experience.

The tag resource allows you to configure and create a tag at Anexia Cloud.

## Example Usage

```hcl
resource "anxcloud_tag" "example" {
  name       = "tag-name"
  service_id = "123asi0uj4398j23i23j41231asd"
}
```

## Argument Reference

- `name` - (Required) The tag name.
- `service_id` - (Optional) The identifier of the service this tag should be assigned to.
- `customer_id` - (Optional) The identifier of the customer this tag should be assigned to. Leave empty to assign to the organization of the logged in user. 

## Attributes Reference

In addition to all the arguments above, the following attributes are exported:

- `organisation_assignments` - Organisation assignments. See [organisation assignments](#organisation-assignments) below for details.

### Organisation assignments

- `customer` - Customer related info. See [customer](#customer) below for details.
- `service` - Service related info. See [service](#service) below for details.

### Customer

- `id` - Tag identifier.
- `customer_id` - Cusomter identifier.
- `demo` - Whether is demo.
- `name` - Customer name.
- `name_slug` - Customer slug name.
- `reseller` - Reseller name.
 
### Service

- `id` - Service identifier.
- `name` - Service name.
