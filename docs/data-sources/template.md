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
- `params` - Template parameters. See [Params](#params) below for details.

### Params

- `hostname` - The requirements for hostname field.
- `cpus` - The requirements for cpus field.
- `memory_mb` - The requirements for memory mb field.
- `disk_gb` - The requirements for disk gb field.
- `dns0` - The requirements for dns0 field.
- `dns1` - The requirements for dns1 field.
- `dns2` - The requirements for dns2 field.
- `dns3` - The requirements for dns3 field.
- `nics` - The requirements for nics field.
- `vlan` - The requirements for vlan field.
- `ips` - The requirements for ips field.
- `boot_delay_seconds` - The requirements for boot delay seconds field.
- `enter_bios_setup` - The requirements for enter bios setup field.
- `password` - The requirements for password field.
- `user` - The requirements for user field.
- `disk_type` - The requirements for disk type field.
