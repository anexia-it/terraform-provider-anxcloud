# DNS Record Import Patterns in Terraform Providers

## Research Summary

This document analyzes how major Terraform providers handle DNS record imports where the record ID alone isn't sufficient to fetch the record from the API.

## Key Findings

### 1. Common Import ID Formats

All major providers use **composite import IDs** that include zone context:

| Provider | Format | Example |
|----------|--------|---------|
| **Cloudflare** | `<zone_id>/<dns_record_id>` | `abc123/def456` |
| **AWS Route53** | `<zone_id>_<name>_<type>_<set_identifier>` | `Z1234_example.com_A_primary` |
| **Google Cloud DNS** | `<project>/<managed_zone>/<name>/<type>` or `<managed_zone>/<name>/<type>` | `my-project/my-zone/example.com./A` |
| **Azure DNS** | Full resource ID | `/subscriptions/.../resourceGroups/.../providers/Microsoft.Network/dnsZones/.../A/...` |

### 2. Implementation Patterns

#### Pattern A: Slash-Separated Format (Cloudflare - Framework)

**Pros:**
- Clean, simple format
- Easy to parse
- URL-encoding support built-in
- Modern terraform-plugin-framework approach

**Implementation:**
```go
func (r *DNSRecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    var data *DNSRecordModel = new(DNSRecordModel)
    
    path_zone_id := ""
    path_dns_record_id := ""
    diags := importpath.ParseImportID(
        req.ID,
        "<zone_id>/<dns_record_id>",
        &path_zone_id,
        &path_dns_record_id,
    )
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }
    
    data.ZoneID = types.StringValue(path_zone_id)
    data.ID = types.StringValue(path_dns_record_id)
    
    // Fetch the record using both zone_id and record_id
    // ... API call ...
    
    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
```

**Helper Function (Cloudflare's importpath package):**
```go
func ParseImportID(str string, format string, args ...any) (diag diag.Diagnostics) {
    path_spec, path := strings.Split(format, "/"), strings.Split(str, "/")
    
    if len(path) != len(path_spec) {
        diag.AddError("invalid ID", fmt.Sprintf("expected urlencoded segments %q, got %q", format, str))
        return
    }
    
    for i, arg := range args {
        segment := path[i]
        switch ptr := arg.(type) {
        case *string:
            *ptr, err = url.PathUnescape(segment)
            // ... error handling ...
        // ... other types ...
        }
    }
    return
}
```

#### Pattern B: Underscore-Separated Format (AWS Route53 - SDK v2)

**Pros:**
- Handles complex identifiers with multiple parts
- Works well with SDK v2
- Supports optional components (set_identifier)

**Implementation:**
```go
func recordParseResourceID(id string) [4]string {
    var recZone, recType, recName, recSet string
    
    parts := strings.Split(id, "_")
    if len(parts) > 1 {
        recZone = parts[0]
    }
    if len(parts) >= 3 {
        // Find record type in parts
        recTypeIndex := -1
        for i, maybeRecType := range parts[1:] {
            if slices.Contains(enum.Values[awstypes.RRType](), maybeRecType) {
                recTypeIndex = i + 1
                break
            }
        }
        if recTypeIndex > 1 {
            recName = strings.Join(parts[1:recTypeIndex], "_")
            recName = strings.TrimSuffix(recName, ".")
            recType = parts[recTypeIndex]
            recSet = strings.Join(parts[recTypeIndex+1:], "_")
        }
    }
    
    return [4]string{recZone, recName, recType, recSet}
}
```

#### Pattern C: Regex-Based Parsing (Google Cloud DNS - SDK v2)

**Pros:**
- Supports multiple import formats
- Flexible for different user inputs
- Can auto-fill project from provider config

**Implementation:**
```go
func resourceDnsRecordSetImportState(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
    config := meta.(*transport_tpg.Config)
    if err := tpgresource.ParseImportId([]string{
        "^projects/(?P<project>[^/]+)/managedZones/(?P<managed_zone>[^/]+)/rrsets/(?P<name>[^/]+)/(?P<type>[^/]+)$",
        "^(?P<project>[^/]+)/(?P<managed_zone>[^/]+)/(?P<name>[^/]+)/(?P<type>[^/]+)$",
        "^(?P<managed_zone>[^/]+)/(?P<name>[^/]+)/(?P<type>[^/]+)$",
    }, d, config); err != nil {
        return nil, err
    }
    
    // Construct full resource ID
    id, err := tpgresource.ReplaceVars(d, config, "projects/{{project}}/managedZones/{{managed_zone}}/rrsets/{{name}}/{{type}}")
    if err != nil {
        return nil, fmt.Errorf("Error constructing id: %s", err)
    }
    d.SetId(id)
    
    return []*schema.ResourceData{d}, nil
}
```

#### Pattern D: Full Resource ID (Azure DNS - SDK v2)

**Pros:**
- Globally unique identifier
- Consistent with Azure's resource model
- Built-in validation

**Implementation:**
```go
Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
    parsed, err := recordsets.ParseRecordTypeID(id)
    if err != nil {
        return err
    }
    if parsed.RecordType != recordsets.RecordTypeA {
        return fmt.Errorf("this resource only supports 'A' records")
    }
    return nil
}),
```

### 3. ImportStatePassthroughContext vs Custom StateContext

#### ImportStatePassthroughContext
- **Use when:** The resource ID contains all necessary information
- **Behavior:** Simply sets the ID and calls Read
- **Limitation:** Cannot parse composite IDs or set multiple attributes

```go
Importer: &schema.ResourceImporter{
    StateContext: schema.ImportStatePassthroughContext,
},
```

#### Custom StateContext Function
- **Use when:** Need to parse composite IDs or set multiple attributes
- **Behavior:** Custom logic to parse ID and populate state
- **Flexibility:** Can validate format, set multiple fields, call API

```go
Importer: &schema.ResourceImporter{
    StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
        // Parse composite ID
        parts := strings.Split(d.Id(), "/")
        if len(parts) != 2 {
            return nil, fmt.Errorf("invalid import ID format, expected: zone_name/record_id")
        }
        
        zoneName := parts[0]
        recordID := parts[1]
        
        // Set attributes needed for Read
        d.Set("zone_name", zoneName)
        d.SetId(recordID)
        
        return []*schema.ResourceData{d}, nil
    },
},
```

### 4. Best Practices for CloudDNS Records

Based on the research, here are recommendations for the anxcloud provider:

#### Recommended Approach: Slash-Separated Format

**Format:** `<zone_name>/<record_identifier>`

**Rationale:**
1. **Consistency:** Matches Cloudflare and modern provider patterns
2. **Simplicity:** Easy for users to understand and construct
3. **URL-safe:** Can handle special characters via URL encoding
4. **Framework-ready:** Works well with terraform-plugin-framework migration

#### Implementation for anxcloud

**Option 1: SDK v2 Custom Import (Current Provider)**

```go
func resourceDNSRecord() *schema.Resource {
    return &schema.Resource{
        // ... other fields ...
        Importer: &schema.ResourceImporter{
            StateContext: resourceDNSRecordImport,
        },
    }
}

func resourceDNSRecordImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
    // Parse import ID: zone_name/record_identifier
    parts := strings.Split(d.Id(), "/")
    if len(parts) != 2 {
        return nil, fmt.Errorf("invalid import ID format, expected: <zone_name>/<record_identifier>, got: %s", d.Id())
    }
    
    zoneName := parts[0]
    recordIdentifier := parts[1]
    
    // Set zone_name so Read can use it
    if err := d.Set("zone_name", zoneName); err != nil {
        return nil, err
    }
    
    // Set the ID to the record identifier
    d.SetId(recordIdentifier)
    
    // Now call Read to populate the rest
    return []*schema.ResourceData{d}, nil
}
```

**Option 2: Alternative Format with All Record Details**

For cases where the API doesn't support lookup by identifier alone:

**Format:** `<zone_name>/<name>/<type>/<rdata>`

```go
func resourceDNSRecordImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
    parts := strings.Split(d.Id(), "/")
    if len(parts) < 3 {
        return nil, fmt.Errorf("invalid import ID format, expected: <zone_name>/<name>/<type>[/<rdata>], got: %s", d.Id())
    }
    
    zoneName := parts[0]
    name := parts[1]
    recordType := parts[2]
    
    // Set required fields
    d.Set("zone_name", zoneName)
    d.Set("name", name)
    d.Set("type", recordType)
    
    // If rdata provided, set it too
    if len(parts) == 4 {
        rdata, _ := url.QueryUnescape(parts[3])
        d.Set("rdata", rdata)
    }
    
    // Generate a temporary ID for Read to find the record
    d.SetId(resourceDNSRecordCanonicalIdentifier(dnsRecordFromResourceData(d)))
    
    return []*schema.ResourceData{d}, nil
}
```

**Option 3: Framework-based (Future Migration)**

```go
func (r *DNSRecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    parts := strings.Split(req.ID, "/")
    if len(parts) != 2 {
        resp.Diagnostics.AddError(
            "Invalid Import ID",
            fmt.Sprintf("Expected format: <zone_name>/<record_identifier>, got: %s", req.ID),
        )
        return
    }
    
    zoneName := parts[0]
    recordIdentifier := parts[1]
    
    // Set the state
    resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("zone_name"), zoneName)...)
    resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), recordIdentifier)...)
}
```

### 5. Documentation Examples

**User-facing documentation:**

```markdown
## Import

DNS records can be imported using the zone name and record identifier, separated by a forward slash:

```shell
terraform import anxcloud_dns_record.example example.com/abc123def456
```

Where:
- `example.com` is the zone name
- `abc123def456` is the record identifier

You can find the record identifier in the Anexia Cloud Portal or by using the data source:

```hcl
data "anxcloud_dns_records" "example" {
  zone_name = "example.com"
}
```
```

### 6. Testing Import

```go
func TestAccAnxCloudDNSRecord_import(t *testing.T) {
    resourceName := "anxcloud_dns_record.test"
    zoneName := "example.com"
    
    resource.Test(t, resource.TestCase{
        PreCheck:     func() { testAccPreCheck(t) },
        Providers:    testAccProviders,
        CheckDestroy: testAccCheckAnxCloudDNSRecordDestroy,
        Steps: []resource.TestStep{
            {
                Config: testAccAnxCloudDNSRecordConfig_basic(zoneName),
                Check: resource.ComposeTestCheckFunc(
                    testAccCheckAnxCloudDNSRecordExists(resourceName),
                ),
            },
            {
                ResourceName:      resourceName,
                ImportState:       true,
                ImportStateVerify: true,
                ImportStateIdFunc: func(s *terraform.State) (string, error) {
                    rs, ok := s.RootModule().Resources[resourceName]
                    if !ok {
                        return "", fmt.Errorf("Not found: %s", resourceName)
                    }
                    
                    zoneName := rs.Primary.Attributes["zone_name"]
                    id := rs.Primary.ID
                    
                    return fmt.Sprintf("%s/%s", zoneName, id), nil
                },
            },
        },
    })
}
```

## Conclusion

For the anxcloud provider's CloudDNS records:

1. **Use slash-separated format:** `<zone_name>/<record_identifier>`
2. **Implement custom StateContext** to parse the composite ID
3. **Set zone_name** before calling Read
4. **Validate format** and provide clear error messages
5. **Document the format** with examples
6. **Add import tests** to ensure it works correctly

This approach:
- ✅ Matches industry standards (Cloudflare, Google)
- ✅ Is user-friendly and intuitive
- ✅ Works with current SDK v2 implementation
- ✅ Is compatible with future Framework migration
- ✅ Handles the zone context requirement elegantly
