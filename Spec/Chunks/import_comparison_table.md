# DNS Record Import - Provider Comparison

## Quick Reference Table

| Provider | Import Format | Example | Separator | Complexity |
|----------|--------------|---------|-----------|------------|
| **Cloudflare** | `<zone_id>/<record_id>` | `abc123/def456` | `/` | ⭐ Simple |
| **Google Cloud DNS** | `<zone>/<name>/<type>` | `my-zone/example.com./A` | `/` | ⭐⭐ Medium |
| **AWS Route53** | `<zone>_<name>_<type>_<set>` | `Z123_example.com_A_primary` | `_` | ⭐⭐⭐ Complex |
| **Azure DNS** | Full ARM resource ID | `/subscriptions/.../A/record` | `/` | ⭐⭐⭐⭐ Very Complex |
| **anxcloud (current)** | `<identifier>` | `abc123` | N/A | ❌ **Broken** |
| **anxcloud (recommended)** | `<zone_name>/<identifier>` | `example.com/abc123` | `/` | ⭐ Simple |

## Implementation Patterns

### Pattern 1: Slash-Separated (Recommended for anxcloud)

**Used by:** Cloudflare, Google Cloud DNS, Azure DNS

**Pros:**
- ✅ Industry standard
- ✅ URL-safe with encoding
- ✅ Easy to parse
- ✅ Human-readable
- ✅ Works with terraform-plugin-framework

**Cons:**
- ⚠️ Requires URL encoding for special characters

**Example Implementation:**
```go
parts := strings.Split(d.Id(), "/")
zoneName := parts[0]
recordID := parts[1]
```

### Pattern 2: Underscore-Separated

**Used by:** AWS Route53

**Pros:**
- ✅ No encoding needed
- ✅ Handles complex multi-part IDs

**Cons:**
- ❌ More complex parsing
- ❌ Ambiguous if values contain underscores
- ❌ Less common pattern

**Example Implementation:**
```go
parts := strings.Split(id, "_")
// Complex logic to find record type in parts
```

### Pattern 3: Full Resource ID

**Used by:** Azure DNS

**Pros:**
- ✅ Globally unique
- ✅ Consistent with platform

**Cons:**
- ❌ Very long and complex
- ❌ Hard to construct manually
- ❌ Platform-specific

## Code Comparison

### Current (Broken)

```go
Importer: &schema.ResourceImporter{
    StateContext: schema.ImportStatePassthroughContext,
},
```

**Usage:**
```bash
terraform import anxcloud_dns_record.example abc123
```

**Result:** ❌ Fails - no zone context

### Recommended (Working)

```go
Importer: &schema.ResourceImporter{
    StateContext: resourceDNSRecordImport,
},

func resourceDNSRecordImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
    parts := strings.Split(d.Id(), "/")
    if len(parts) != 2 {
        return nil, fmt.Errorf("invalid format")
    }
    d.Set("zone_name", parts[0])
    d.SetId(parts[1])
    return []*schema.ResourceData{d}, nil
}
```

**Usage:**
```bash
terraform import anxcloud_dns_record.example example.com/abc123
```

**Result:** ✅ Works - has zone context

## Real-World Examples

### Cloudflare DNS Record

```bash
# Import format
terraform import cloudflare_dns_record.example 023e105f4ecef8ad9ca31a8372d0c353/372e67954025e0ba6aaa6d586b9e0b59

# Where:
# - 023e105f4ecef8ad9ca31a8372d0c353 = zone_id
# - 372e67954025e0ba6aaa6d586b9e0b59 = record_id
```

### Google Cloud DNS Record

```bash
# Import format (short)
terraform import google_dns_record_set.example my-zone/example.com./A

# Import format (full)
terraform import google_dns_record_set.example my-project/my-zone/example.com./A

# Where:
# - my-project = project (optional, from provider)
# - my-zone = managed_zone
# - example.com. = name (with trailing dot)
# - A = type
```

### AWS Route53 Record

```bash
# Import format
terraform import aws_route53_record.example Z1234567890ABC_example.com_A_primary

# Where:
# - Z1234567890ABC = zone_id
# - example.com = name
# - A = type
# - primary = set_identifier (optional)
```

### anxcloud (Recommended)

```bash
# Import format
terraform import anxcloud_dns_record.example example.com/abc123def456

# Where:
# - example.com = zone_name
# - abc123def456 = record identifier
```

## Error Handling Comparison

### Cloudflare (Good)

```
Error: invalid ID, expected urlencoded segments "<zone_id>/<dns_record_id>", got "abc123"
```

### Google Cloud DNS (Good)

```
Error: Import ID must match one of:
  - projects/<project>/managedZones/<zone>/rrsets/<name>/<type>
  - <project>/<zone>/<name>/<type>
  - <zone>/<name>/<type>
```

### anxcloud Current (Poor)

```
Error: Cannot import non-existent remote object
```

### anxcloud Recommended (Good)

```
Error: invalid import ID format: expected '<zone_name>/<record_identifier>', got 'abc123'

Example: terraform import anxcloud_dns_record.example example.com/abc123def456
```

## Testing Comparison

### Cloudflare

```go
ImportStateIdFunc: func(s *terraform.State) (string, error) {
    zoneID := s.RootModule().Resources[resourceName].Primary.Attributes["zone_id"]
    id := s.RootModule().Resources[resourceName].Primary.ID
    return fmt.Sprintf("%s/%s", zoneID, id), nil
}
```

### Google Cloud DNS

```go
ImportStateIdFunc: func(s *terraform.State) (string, error) {
    zone := s.RootModule().Resources[resourceName].Primary.Attributes["managed_zone"]
    name := s.RootModule().Resources[resourceName].Primary.Attributes["name"]
    recordType := s.RootModule().Resources[resourceName].Primary.Attributes["type"]
    return fmt.Sprintf("%s/%s/%s", zone, name, recordType), nil
}
```

### anxcloud Recommended

```go
ImportStateIdFunc: func(s *terraform.State) (string, error) {
    zoneName := s.RootModule().Resources[resourceName].Primary.Attributes["zone_name"]
    id := s.RootModule().Resources[resourceName].Primary.ID
    return fmt.Sprintf("%s/%s", zoneName, id), nil
}
```

## Decision Matrix

| Criteria | Slash-Separated | Underscore-Separated | Full Resource ID |
|----------|----------------|---------------------|------------------|
| **Simplicity** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐ |
| **Readability** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐ |
| **Industry Standard** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐ |
| **URL Safety** | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ |
| **Framework Compatible** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ |
| **Error Messages** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ |
| **User Experience** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐ |

**Winner:** Slash-Separated Format ✅

## Recommendation

**Use slash-separated format: `<zone_name>/<record_identifier>`**

**Reasons:**
1. ✅ Matches Cloudflare (most similar DNS provider)
2. ✅ Simple and intuitive
3. ✅ Industry standard
4. ✅ Easy to implement
5. ✅ Great error messages
6. ✅ Future-proof

**Implementation:** See `anxcloud_dns_import_recommendation.md`
