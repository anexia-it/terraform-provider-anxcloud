---
page_title: "template Data Source - terraform-provider-anxcloud"
subcategory: ""
description: |-
  The template data source allows you to retrieve information about available templates for specified location.
---

# Data Source `anxcloud_template`

The template data source allows you to retrieve information about available templates for specified location.

## Example Usage

```hcl
data "anxcloud_template" "example" {
  location_id = "52b5f6b2fd3a4a7eaaedf1a7c019e9ea"
}
```

## Argument Reference

- `location_id` - (Required) Location identifier.
- `template_type` - (Optional) Template type, either 'templates' or 'from_scratch'. Defaults to 'templates'.

## Attributes Reference

The following attributes are exported.

- `id` - Template identifier.
- `name` - Operating system name defined in the template.
- `bit` - Operating system word size.
- `build` - Operating system build.
