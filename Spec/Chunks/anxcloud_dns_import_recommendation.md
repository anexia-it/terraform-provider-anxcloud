# DNS Record Import Implementation for anxcloud Provider

## Executive Summary

Based on research of major Terraform providers (Cloudflare, AWS Route53, Google Cloud DNS, Azure DNS), here's the recommended implementation for anxcloud DNS record imports.

## Problem Statement

The anxcloud CloudDNS API requires both:
1. **Zone name** (zone context)
2. **Record identifier** (unique ID within the zone)

Currently, `ImportStatePassthroughContext` only passes the ID, which is insufficient.

## Recommended Solution

### Import Format

Use **slash-separated format**: `<zone_name>/<record_identifier>`

**Example:**
```bash
terraform import anxcloud_dns_record.example example.com/abc123def456
```

### Implementation

Replace the current import configuration:

```go
// BEFORE (doesn't work properly)
Importer: &schema.ResourceImporter{
    StateContext: schema.ImportStatePassthroughContext,
},
```

With a custom import function:

```go
// AFTER (works correctly)
Importer: &schema.ResourceImporter{
    StateContext: resourceDNSRecordImport,
},
```

### Complete Implementation Code

Add this function to `anxcloud/resource_dns_record.go`:

```go
func resourceDNSRecordImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
    // Parse import ID format: zone_name/record_identifier
    parts := strings.Split(d.Id(), "/")
    if len(parts) != 2 {
        return nil, fmt.Errorf(
            "invalid import ID format: expected '<zone_name>/<record_identifier>', got '%s'\n\n"+
                "Example: terraform import anxcloud_dns_record.example example.com/abc123def456",
            d.Id(),
        )
    }
    
    zoneName := parts[0]
    recordIdentifier := parts[1]
    
    // Validate zone name is not empty
    if zoneName == "" {
        return nil, fmt.Errorf("zone_name cannot be empty in import ID")
    }
    
    // Validate record identifier is not empty
    if recordIdentifier == "" {
        return nil, fmt.Errorf("record_identifier cannot be empty in import ID")
    }
    
    // Set zone_name attribute so Read can use it
    if err := d.Set("zone_name", zoneName); err != nil {
        return nil, fmt.Errorf("failed to set zone_name: %w", err)
    }
    
    // Set the ID to the record identifier
    d.SetId(recordIdentifier)
    
    // The Read function will be called automatically after this
    // and will populate all other attributes
    return []*schema.ResourceData{d}, nil
}
```

### Update Resource Definition

In `resourceDNSRecord()` function:

```go
func resourceDNSRecord() *schema.Resource {
    return &schema.Resource{
        Description: "This resource allows you to create DNS records for a specified zone. TXT records might behave funny, we are working on it." +
            " Create and delete operations will be handled in batches internally. As a side effect this will cause whole batches to fail in case some of the operations are invalid." +
            " Updating record attributes triggers a replacement (destroy old -> create new).",
        CreateContext: resourceDNSRecordCreate,
        ReadContext:   resourceDNSRecordRead,
        DeleteContext: resourceDNSRecordDelete,
        Importer: &schema.ResourceImporter{
            StateContext: resourceDNSRecordImport,  // Changed from ImportStatePassthroughContext
        },
        Timeouts: &schema.ResourceTimeout{
            Create: schema.DefaultTimeout(2 * time.Minute),
            Read:   schema.DefaultTimeout(time.Minute),
            Delete: schema.DefaultTimeout(2 * time.Minute),
        },
        Schema: schemaDNSRecord(),
    }
}
```

## Documentation Update

Update `docs/resources/dns_record.md` to include import section:

```markdown
## Import

DNS records can be imported using the zone name and record identifier, separated by a forward slash:

```shell
terraform import anxcloud_dns_record.example example.com/abc123def456
```

Where:
- `example.com` is the zone name
- `abc123def456` is the record identifier

### Finding the Record Identifier

You can find the record identifier using the `anxcloud_dns_records` data source:

```hcl
data "anxcloud_dns_records" "example" {
  zone_name = "example.com"
}

output "record_identifiers" {
  value = {
    for record in data.anxcloud_dns_records.example.records :
    "${record.name}.${record.type}" => record.identifier
  }
}
```

Or by inspecting an existing Terraform state:

```shell
terraform state show anxcloud_dns_record.example | grep "^id"
```
```

## Testing

Add import test to `anxcloud/resource_dns_record_test.go`:

```go
func TestAccAnxCloudDNSRecord_import(t *testing.T) {
    resourceName := "anxcloud_dns_record.test"
    zoneName := testAccAnxCloudDNSZoneName()
    
    resource.ParallelTest(t, resource.TestCase{
        PreCheck:          func() { testAccPreCheck(t) },
        ProviderFactories: testAccProviderFactories,
        CheckDestroy:      testAccCheckAnxCloudDNSRecordDestroy,
        Steps: []resource.TestStep{
            {
                Config: testAccAnxCloudDNSRecordConfig_basic(zoneName, "test", "A", "192.0.2.1", 3600),
                Check: resource.ComposeTestCheckFunc(
                    testAccCheckAnxCloudDNSRecordExists(resourceName),
                    resource.TestCheckResourceAttr(resourceName, "zone_name", zoneName),
                    resource.TestCheckResourceAttr(resourceName, "name", "test"),
                    resource.TestCheckResourceAttr(resourceName, "type", "A"),
                ),
            },
            {
                ResourceName:      resourceName,
                ImportState:       true,
                ImportStateVerify: true,
                ImportStateIdFunc: testAccAnxCloudDNSRecordImportStateIdFunc(resourceName),
            },
        },
    })
}

func testAccAnxCloudDNSRecordImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
    return func(s *terraform.State) (string, error) {
        rs, ok := s.RootModule().Resources[resourceName]
        if !ok {
            return "", fmt.Errorf("resource not found: %s", resourceName)
        }
        
        zoneName := rs.Primary.Attributes["zone_name"]
        if zoneName == "" {
            return "", fmt.Errorf("zone_name not set in resource %s", resourceName)
        }
        
        id := rs.Primary.ID
        if id == "" {
            return "", fmt.Errorf("ID not set in resource %s", resourceName)
        }
        
        // Return the composite import ID
        return fmt.Sprintf("%s/%s", zoneName, id), nil
    }
}

func TestAccAnxCloudDNSRecord_importInvalidFormat(t *testing.T) {
    resource.ParallelTest(t, resource.TestCase{
        PreCheck:          func() { testAccPreCheck(t) },
        ProviderFactories: testAccProviderFactories,
        Steps: []resource.TestStep{
            {
                Config:            testAccAnxCloudDNSRecordConfig_basic("example.com", "test", "A", "192.0.2.1", 3600),
                ResourceName:      "anxcloud_dns_record.test",
                ImportState:       true,
                ImportStateId:     "invalid-format-no-slash",
                ExpectError:       regexp.MustCompile("invalid import ID format"),
            },
        },
    })
}
```

## Migration Guide for Users

If users have existing imported resources with the old format, they'll need to re-import:

```bash
# 1. Remove from state (doesn't delete the actual resource)
terraform state rm anxcloud_dns_record.example

# 2. Re-import with new format
terraform import anxcloud_dns_record.example example.com/abc123def456

# 3. Verify
terraform plan  # Should show no changes
```

## Alternative: Support Multiple Formats

If you want to be more flexible and support both identifier-only (for backward compatibility) and the new format:

```go
func resourceDNSRecordImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
    parts := strings.Split(d.Id(), "/")
    
    switch len(parts) {
    case 1:
        // Legacy format: just the identifier
        // User must have zone_name in their config
        recordIdentifier := parts[0]
        d.SetId(recordIdentifier)
        
        // Note: zone_name must be in the config for this to work
        // The Read function will fail if zone_name is not set
        return []*schema.ResourceData{d}, nil
        
    case 2:
        // New format: zone_name/record_identifier
        zoneName := parts[0]
        recordIdentifier := parts[1]
        
        if zoneName == "" || recordIdentifier == "" {
            return nil, fmt.Errorf("zone_name and record_identifier cannot be empty")
        }
        
        if err := d.Set("zone_name", zoneName); err != nil {
            return nil, fmt.Errorf("failed to set zone_name: %w", err)
        }
        
        d.SetId(recordIdentifier)
        return []*schema.ResourceData{d}, nil
        
    default:
        return nil, fmt.Errorf(
            "invalid import ID format: expected '<zone_name>/<record_identifier>' or '<record_identifier>', got '%s'\n\n"+
                "Recommended format: terraform import anxcloud_dns_record.example example.com/abc123def456",
            d.Id(),
        )
    }
}
```

## Benefits

1. ✅ **Works correctly** - Provides zone context needed by the API
2. ✅ **User-friendly** - Clear, intuitive format
3. ✅ **Industry standard** - Matches Cloudflare and Google patterns
4. ✅ **Good error messages** - Helps users understand the correct format
5. ✅ **Testable** - Easy to write comprehensive tests
6. ✅ **Future-proof** - Compatible with terraform-plugin-framework migration

## Comparison with Current Implementation

| Aspect | Current (Passthrough) | Recommended (Custom) |
|--------|----------------------|---------------------|
| **Works?** | ❌ No - missing zone context | ✅ Yes - includes zone context |
| **Format** | `abc123def456` | `example.com/abc123def456` |
| **User Experience** | ❌ Confusing errors | ✅ Clear error messages |
| **Documentation** | ❌ Unclear how to import | ✅ Clear examples |
| **Testing** | ❌ Can't test properly | ✅ Comprehensive tests |

## Implementation Checklist

- [ ] Add `resourceDNSRecordImport` function to `resource_dns_record.go`
- [ ] Update `Importer` in `resourceDNSRecord()` function
- [ ] Add import documentation to `docs/resources/dns_record.md`
- [ ] Add import tests to `resource_dns_record_test.go`
- [ ] Test manually with real resources
- [ ] Update CHANGELOG.md with breaking change note (if not supporting legacy format)
- [ ] Consider adding migration guide in documentation

## Breaking Change Considerations

If you want to avoid breaking existing users (if any are using import):

1. **Support both formats** (recommended for transition period)
2. **Add deprecation warning** for old format
3. **Remove old format** in next major version

Or, if import is not widely used yet:

1. **Just implement the new format** (simpler, cleaner)
2. **Document the change** in CHANGELOG
3. **Provide migration instructions**
