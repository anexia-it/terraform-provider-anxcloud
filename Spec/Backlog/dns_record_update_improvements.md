# DNS Record Update Improvements - Backlog Item

## Problem Description

Currently, all DNS record field modifications trigger complete resource replacement (destroy + create) instead of in-place updates. This occurs because the schema marks all fields as `ForceNew: true`, even fields that should be updatable according to DNS standards and CloudDNS API capabilities.

## Current State Analysis

### Schema Issues
The current `schema_dns_record.go` incorrectly marks these fields as `ForceNew: true`:
- `rdata` (record data) - should be updatable
- `ttl` (time-to-live) - should be updatable

Fields correctly marked as `ForceNew: true`:
- `name` - domain name (DNS fundamental)
- `type` - record type (DNS fundamental)
- `zone_name` - zone scope (major change)

### Impact on Users
- **TTL updates** require record replacement with downtime
- **RDATA changes** (IP address updates, etc.) require replacement
- **Poor user experience** for common DNS maintenance tasks
- **Unnecessary API calls** and resource churn

## DNS Standards & API Research

### RFC 2136 (DNS UPDATE) Compliance
- **TTL and RDATA** can be updated in-place without replacement
- **NAME and TYPE changes** require delete/add operations (replacement)
- **CLASS changes** are rare and typically require replacement

### CloudDNS API Capabilities
- **Update operations supported** via `apiClient.Update()`
- **Stable identifiers** persist after updates (true in-place updates)
- **Immutable field** indicates some records may be protected
- **Current implementation** uses batch changesets but doesn't leverage updates

### Provider Best Practices
- **AWS Route53**: Allows in-place updates for most fields
- **Google Cloud DNS**: Similar in-place update support
- **Cloudflare**: More restrictive but still allows some in-place updates

## Proposed Solution

### Phase 1: Schema Updates
Remove `ForceNew: true` from updatable fields:

```go
"ttl": {
    Type:        schema.TypeInt,
    Optional:    true,
    // Remove ForceNew: true
    Description: "Region specific TTL. If not set the zone TTL will be used.",
},
"rdata": {
    Type:        schema.TypeString,
    Required:    true,
    // Remove ForceNew: true
    Description: "DNS record data.",
},
```

### Phase 2: Update Function Implementation
Implement `resourceDNSRecordUpdate()` function:

```go
func resourceDNSRecordUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
    // Check if record is immutable
    if d.Get("immutable").(bool) {
        return diag.Errorf("cannot update immutable DNS record")
    }

    // Use CloudDNS API Update instead of destroy/create
    a := apiFromProviderConfig(m)
    record := dnsRecordFromResourceData(d)

    if err := a.Update(ctx, &record); err != nil {
        return diag.FromErr(err)
    }

    return resourceDNSRecordRead(ctx, d, m)
}
```

### Phase 3: Resource Definition Updates
Add UpdateContext to resource definition:

```go
return &schema.Resource{
    // ... existing fields ...
    UpdateContext: resourceDNSRecordUpdate,
    // ... rest ...
}
```

## Implementation Requirements

### Code Changes
1. **Schema Updates**: Remove ForceNew from TTL and RDATA fields
2. **Update Function**: Implement in-place update logic
3. **Error Handling**: Handle immutable records appropriately
4. **API Integration**: Use CloudDNS Update API instead of changeset operations

### Testing Requirements
1. **Unit Tests**: Test update function with various field changes
2. **Integration Tests**: Test TTL and RDATA updates via CloudDNS API
3. **Regression Tests**: Ensure ForceNew fields still trigger replacement
4. **Immutable Handling**: Test behavior with immutable records

### Documentation Updates
1. **Schema Documentation**: Update field descriptions to indicate updatability
2. **Resource Documentation**: Document which fields support in-place updates
3. **Migration Guide**: Explain behavior changes for users

## Risk Assessment

### Low Risk Changes
- **TTL updates**: Standard DNS operation, widely supported
- **RDATA updates**: Standard DNS operation, widely supported
- **API compatibility**: CloudDNS Update API already exists

### Backward Compatibility
- **✅ Safe**: Existing configurations continue to work
- **✅ Non-breaking**: No changes to existing behavior for ForceNew fields
- **✅ Progressive**: Adds new capabilities without removing old ones

### Rollback Plan
- **Easy rollback**: Can restore ForceNew flags if issues arise
- **Feature flag**: Could add provider-level flag to disable updates
- **Gradual rollout**: Can be deployed incrementally

## Success Criteria

- ✅ **TTL updates** happen in-place without replacement
- ✅ **RDATA updates** happen in-place without replacement
- ✅ **ForceNew fields** still trigger replacement as expected
- ✅ **Immutable records** are handled appropriately
- ✅ **API calls** are more efficient (update vs destroy+create)
- ✅ **User experience** improved for common DNS maintenance
- ✅ **RFC 2136 compliance** achieved

## Estimated Timeline

- **Phase 1 (Schema)**: 1-2 hours
- **Phase 2 (Update Function)**: 2-4 hours
- **Phase 3 (Testing)**: 4-6 hours
- **Total**: 1-2 days

## Dependencies

- CloudDNS API Update functionality (already exists)
- Testing environment access
- Review of immutable record handling
- Documentation updates

## Business Value

- **Improved User Experience**: Common DNS operations no longer require replacement
- **Better Performance**: In-place updates are faster and more reliable
- **Standards Compliance**: Aligns with DNS protocol specifications
- **Reduced API Load**: Fewer destroy/create cycles on CloudDNS API
- **Competitive Parity**: Matches capabilities of other DNS Terraform providers

This change will significantly improve the user experience for DNS record management while maintaining full backward compatibility and following DNS industry standards.</content>
<parameter name="filePath">Spec/Backlog/dns_record_update_improvements.md