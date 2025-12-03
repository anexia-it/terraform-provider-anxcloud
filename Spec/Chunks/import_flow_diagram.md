# DNS Record Import Flow Diagram

## Current Flow (Broken)

```
User runs:
terraform import anxcloud_dns_record.example abc123def456
                                              ↓
                        ImportStatePassthroughContext
                                              ↓
                                    d.SetId("abc123def456")
                                              ↓
                                    resourceDNSRecordRead()
                                              ↓
                        Tries to read record with ID only
                                              ↓
                        ❌ FAILS - No zone_name context!
                                              ↓
                        Error: Cannot find record
```

## Recommended Flow (Working)

```
User runs:
terraform import anxcloud_dns_record.example example.com/abc123def456
                                              ↓
                        resourceDNSRecordImport()
                                              ↓
                        Parse ID: "example.com/abc123def456"
                                              ↓
                        Split by "/" → ["example.com", "abc123def456"]
                                              ↓
                        Validate: len(parts) == 2 ✓
                                              ↓
                        d.Set("zone_name", "example.com")
                        d.SetId("abc123def456")
                                              ↓
                        Return []*schema.ResourceData{d}, nil
                                              ↓
                        Terraform automatically calls Read
                                              ↓
                        resourceDNSRecordRead()
                                              ↓
                        Has both zone_name AND identifier!
                                              ↓
                        ✅ SUCCESS - Record found and imported
```

## Code Flow Comparison

### Current (Broken)

```go
// 1. User imports
$ terraform import anxcloud_dns_record.example abc123

// 2. Terraform calls
Importer.StateContext = schema.ImportStatePassthroughContext

// 3. PassthroughContext does
d.SetId("abc123")
// zone_name is NOT set!

// 4. Read is called
func resourceDNSRecordRead(ctx, d, m) {
    zoneName := d.Get("zone_name").(string)  // ❌ Empty!
    recordID := d.Id()                        // ✓ "abc123"
    
    // API call needs BOTH
    findDNSRecord(ctx, api, Record{
        ZoneName: zoneName,    // ❌ Empty string!
        Identifier: recordID,  // ✓ "abc123"
    })
    // ❌ FAILS - zone_name is required
}
```

### Recommended (Working)

```go
// 1. User imports
$ terraform import anxcloud_dns_record.example example.com/abc123

// 2. Terraform calls
Importer.StateContext = resourceDNSRecordImport

// 3. Custom import function
func resourceDNSRecordImport(ctx, d, m) {
    parts := strings.Split(d.Id(), "/")
    // parts = ["example.com", "abc123"]
    
    if len(parts) != 2 {
        return nil, fmt.Errorf("invalid format")
    }
    
    d.Set("zone_name", parts[0])  // ✓ "example.com"
    d.SetId(parts[1])              // ✓ "abc123"
    
    return []*schema.ResourceData{d}, nil
}

// 4. Read is called
func resourceDNSRecordRead(ctx, d, m) {
    zoneName := d.Get("zone_name").(string)  // ✓ "example.com"
    recordID := d.Id()                        // ✓ "abc123"
    
    // API call has BOTH
    findDNSRecord(ctx, api, Record{
        ZoneName: zoneName,    // ✓ "example.com"
        Identifier: recordID,  // ✓ "abc123"
    })
    // ✅ SUCCESS - record found!
}
```

## State Comparison

### Before Import (Empty State)

```
State: {}
```

### After Import - Current (Broken)

```
State: {
  "id": "abc123def456",
  "zone_name": "",           ❌ Missing!
  "name": "",                ❌ Missing!
  "type": "",                ❌ Missing!
  "rdata": "",               ❌ Missing!
  // ... all other fields empty
}

Error: Cannot import - zone_name required
```

### After Import - Recommended (Working)

```
State: {
  "id": "abc123def456",
  "zone_name": "example.com",  ✓ Set by import
  "name": "www",               ✓ Populated by Read
  "type": "A",                 ✓ Populated by Read
  "rdata": "192.0.2.1",        ✓ Populated by Read
  "ttl": 3600,                 ✓ Populated by Read
  // ... all fields populated
}

Success: Record imported successfully
```

## Error Handling Flow

### Current (Poor UX)

```
User: terraform import anxcloud_dns_record.example abc123
                                              ↓
                        ImportStatePassthroughContext
                                              ↓
                                    Read() called
                                              ↓
                        API error: zone_name required
                                              ↓
                        Generic error message:
                        "Error: Cannot import non-existent remote object"
                                              ↓
                        ❌ User confused - what went wrong?
```

### Recommended (Good UX)

```
User: terraform import anxcloud_dns_record.example abc123
                                              ↓
                        resourceDNSRecordImport()
                                              ↓
                        Parse: "abc123" → split by "/"
                                              ↓
                        parts = ["abc123"] (length 1, not 2)
                                              ↓
                        Validation fails
                                              ↓
                        Clear error message:
                        "Error: invalid import ID format: expected 
                        '<zone_name>/<record_identifier>', got 'abc123'
                        
                        Example: terraform import anxcloud_dns_record.example 
                        example.com/abc123def456"
                                              ↓
                        ✅ User knows exactly what to do!
```

## Data Flow Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                    User Input                                │
│  terraform import anxcloud_dns_record.example                │
│                   example.com/abc123def456                   │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│              resourceDNSRecordImport()                       │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ 1. Parse: strings.Split(id, "/")                    │   │
│  │    → ["example.com", "abc123def456"]                │   │
│  │                                                       │   │
│  │ 2. Validate: len(parts) == 2 ✓                      │   │
│  │                                                       │   │
│  │ 3. Set zone_name: d.Set("zone_name", "example.com") │   │
│  │                                                       │   │
│  │ 4. Set ID: d.SetId("abc123def456")                  │   │
│  └─────────────────────────────────────────────────────┘   │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                 Terraform Framework                          │
│  Automatically calls resourceDNSRecordRead()                 │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│              resourceDNSRecordRead()                         │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ 1. Get zone_name: "example.com" ✓                   │   │
│  │                                                       │   │
│  │ 2. Get ID: "abc123def456" ✓                         │   │
│  │                                                       │   │
│  │ 3. Call API: findDNSRecord(zone, id)                │   │
│  │                                                       │   │
│  │ 4. Populate all attributes from API response         │   │
│  └─────────────────────────────────────────────────────┘   │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                   Terraform State                            │
│  {                                                           │
│    "id": "abc123def456",                                     │
│    "zone_name": "example.com",                               │
│    "name": "www",                                            │
│    "type": "A",                                              │
│    "rdata": "192.0.2.1",                                     │
│    "ttl": 3600,                                              │
│    ...                                                       │
│  }                                                           │
└─────────────────────────────────────────────────────────────┘
```

## Comparison with Other Providers

### Cloudflare (Similar Pattern)

```
Import: zone_id/record_id
        ↓
Parse: ["zone_id", "record_id"]
        ↓
Set: zone_id ✓, id ✓
        ↓
Read: Has both ✓
        ↓
Success ✅
```

### Google Cloud DNS (More Complex)

```
Import: zone/name/type
        ↓
Parse: ["zone", "name", "type"]
        ↓
Set: managed_zone ✓, name ✓, type ✓
        ↓
Construct ID: "projects/X/managedZones/Y/rrsets/Z/T"
        ↓
Read: Has all context ✓
        ↓
Success ✅
```

### AWS Route53 (Most Complex)

```
Import: zone_id_name_type_set
        ↓
Parse: Complex logic to find type in parts
        ↓
Set: zone_id ✓, name ✓, type ✓, set_identifier ✓
        ↓
Read: Has all context ✓
        ↓
Success ✅
```

### anxcloud (Recommended - Simple)

```
Import: zone_name/record_id
        ↓
Parse: ["zone_name", "record_id"]
        ↓
Set: zone_name ✓, id ✓
        ↓
Read: Has both ✓
        ↓
Success ✅
```

## Key Takeaways

1. **Problem:** PassthroughContext doesn't set zone_name
2. **Solution:** Custom import function parses composite ID
3. **Format:** `zone_name/record_identifier` (simple, standard)
4. **Result:** Read has all required context
5. **Benefit:** Import works correctly ✅

## Implementation Priority

```
High Priority (Must Have):
├── Parse composite ID
├── Validate format
├── Set zone_name
└── Set record ID

Medium Priority (Should Have):
├── Clear error messages
├── Format examples in errors
└── Validation of empty parts

Low Priority (Nice to Have):
├── Support multiple formats
├── URL encoding/decoding
└── Deprecation warnings
```

## Testing Flow

```
1. Create Record
   terraform apply
   ↓
2. Capture State
   terraform state show
   ↓
3. Remove from State
   terraform state rm
   ↓
4. Import
   terraform import zone/id
   ↓
5. Verify
   terraform plan (should show no changes)
   ↓
6. Success ✅
```
