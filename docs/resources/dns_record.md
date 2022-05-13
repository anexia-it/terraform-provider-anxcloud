---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "anxcloud_dns_record Resource - terraform-provider-anxcloud"
subcategory: ""
description: |-
  This resource allows you to create DNS records for a specified zone. TXT records might behave funny, we are working on it.
---

# anxcloud_dns_record (Resource)

This resource allows you to create DNS records for a specified zone. TXT records might behave funny, we are working on it.

## Example Usage

```terraform
resource "anxcloud_dns_record" "example" {
  name      = "webmail"
  zone_name = "example.com"
  type      = "A"
  rdata     = "198.51.100.10"
  ttl       = 3600
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) DNS record name.
- `rdata` (String) DNS record data.
- `type` (String) DNS record type.
- `zone_name` (String) Zone of DNS record.

### Optional

- `id` (String) The ID of this resource.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- `ttl` (Number) Region specific TTL. If not set the zone TTL will be used.

### Read-Only

- `identifier` (String) DNS Record identifier. Changes on revision change and therefore shouldn't be used as reference.
- `immutable` (Boolean) Specifies whether or not a record is immutable.
- `region` (String) DNS record region (for GeoDNS aware records).

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)
- `delete` (String)
- `read` (String)
- `update` (String)

