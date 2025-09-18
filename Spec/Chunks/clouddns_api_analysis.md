# AnxCloud SDK CloudDNS API Analysis

## Overview
This analysis examines the go.anx.io/go-anxcloud v0.9.1-alpha SDK's CloudDNS API implementation, focusing on the clouddns/v1 package and zone.Record structures.

## 1. DNS Record Response Fields

Based on the SDK structures, DNS record responses contain the following fields:

### From `zone.Record` (pkg/clouddns/zone/zone.go):
```go
type Record struct {
    Identifier uuid.UUID `json:"identifier"`
    Immutable  bool      `json:"immutable"`
    Name       string    `json:"name"`
    RData      string    `json:"rdata"`
    Region     string    `json:"region"`
    TTL        *int      `json:"ttl"`
    Type       string    `json:"Type"`  // Note: uppercase T
}
```

### From `v1.Record` (pkg/apis/clouddns/v1/record_types.go):
```go
type Record struct {
    Identifier string `json:"identifier,omitempty" anxcloud:"identifier"`
    ZoneName   string `json:"-"`
    Immutable  bool   `json:"immutable,omitempty"`
    Name       string `json:"name"`
    RData      string `json:"rdata"`
    Region     string `json:"region"`
    TTL        int    `json:"ttl"`
    Type       string `json:"type"`
}
```

**Key differences:**
- `zone.Record` uses `uuid.UUID` for Identifier, `v1.Record` uses `string`
- `zone.Record` has `*int` for TTL (nullable), `v1.Record` has `int`
- `zone.Record` has `Type` with uppercase T in JSON tag, `v1.Record` has `type`

## 2. Identifier Generation and Stability

### Current Implementation Analysis:
- **Identifier Generation**: The Identifier appears to be generated server-side and returned in API responses
- **Stability Issues**: The Terraform provider schema documentation explicitly states:
  ```
  "identifier": {
      Type: schema.TypeString,
      Computed: true,
      Description: "DNS Record identifier. Changes on revision change and therefore shouldn't be used as reference.",
  }
  ```

### Canonical Identifier Approach:
The current Terraform provider uses a content-based canonical identifier created by concatenating:
- Name
- ZoneName  
- Type
- URL-encoded RData
- TTL value
- Region
- Immutable flag

This approach treats records as uniquely identified by their content rather than a stable database ID.

## 3. API Endpoints for Listing/Searching Records

### Primary Endpoints:
- **List Records**: `GET /api/clouddns/v1/zone.json/{zone}/records`
- **Create Record**: `POST /api/clouddns/v1/zone.json/{zone}/records`  
- **Update Record**: `PUT /api/clouddns/v1/zone.json/{zone}/records/{id}`
- **Delete Record**: `DELETE /api/clouddns/v1/zone.json/{zone}/records/{id}`

### List Endpoint Features:
The list endpoint supports query parameters for filtering:
- `name`: Filter by record name
- `data`: Filter by RData content  
- `type`: Filter by record type

### Additional Zone-Level Operations:
- **Apply Changeset**: `POST /api/clouddns/v1/zone.json/{zone}/changeset`
- **Import Zone**: `POST /api/clouddns/v1/zone.json/{zone}/import`

## 4. Unique Keys and Stable Identifiers

### Current Limitations:
1. **Identifier Instability**: The `identifier` field changes with zone revisions, making it unsuitable for long-term resource referencing
2. **Content-Based Identity**: Records are currently identified by their complete content (name, type, data, TTL, region, immutable status)
3. **No Composite Unique Constraints**: The API doesn't appear to enforce unique constraints beyond the content-based approach

### Potential Stable Identifier Candidates:
Based on the code analysis, no clearly stable identifiers exist beyond the current content-based approach. The revision-based nature of DNS zone management means identifiers can change when zones are updated.

### Recommendations:
1. **Continue Content-Based Approach**: The current canonical identifier strategy is appropriate given the API limitations
2. **Monitor API Evolution**: Future SDK versions might introduce stable identifiers
3. **Consider Composite Keys**: If stable identifiers become available, consider migrating to use them for better resource lifecycle management

## Implementation Notes

### TXT Record Handling:
- TXT records require special handling for RData quoting/unquoting
- The engine returns TXT RData enclosed in quotes, but the API expects unquoted input
- The Terraform provider handles this with conditional quoting logic

### Batch Operations:
- Record creation/deletion uses batch operations via changesets
- Multiple operations are grouped and executed together
- Error handling provides per-record error details for batch failures

### Zone State Management:
- Zones have editable/non-editable states
- Operations wait for zones to become editable before proceeding
- This ensures consistency during concurrent operations
