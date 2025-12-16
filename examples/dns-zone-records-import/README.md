# DNS Zone & Records Import Tool

Import existing DNS zones and all their records from Anexia Cloud into Terraform/OpenTofu.

## Prerequisites

- `tofu` (or `terraform`) CLI
- `jq` for JSON parsing
- `curl` for API calls
- `ANEXIA_TOKEN` environment variable set

## Quick Start

### Automatic Mode (Recommended)

```bash
./generate-import-blocks.sh --auto example.com
```

**What it does:**
1. **Fetch zone data** - Queries Anexia Cloud API to get zone identifier and all DNS records
2. **Generate import blocks** - Creates Terraform import configuration with resource stubs
3. **Validate safety** - Ensures the plan contains ONLY imports (no creates/destroys)
4. **Import resources** - Executes `tofu apply` to import into state
5. **Generate config** - Creates final `.tf` files with real values from API
6. **Clean up** - Removes temporary files

**Output:**
- `imported_dns_zone.tf` - Zone resource configuration
- `imported_dns_records.tf` - All DNS record configurations
- `terraform.tfstate` - State with imported resources

### Manual Mode (Step-by-Step)

```bash
./generate-import-blocks.sh example.com
```

**Then execute:**

```bash
# Step 1: Review import blocks
cat _import-blocks.tf

# Step 2: Run plan and save output
# Why: Creates a saved plan file with import operations
tofu plan -out=import.tfplan

# Step 3: Generate resource configuration from the plan
# Why: Terraform needs .tf files with resource definitions
tofu plan -generate-config-out=_generated.tf

# Step 4: Review generated configuration
vi _generated.tf

# Step 5: Import resources into state using saved plan
# Why: Fetches actual resource data from Anexia API into Terraform state
tofu apply import.tfplan

# Step 6: CRITICAL - Save configuration before cleanup!
# Why: cleanup.sh removes _*.tf files including _generated.tf
mv _generated.tf dns_config.tf

# Step 7: Clean up temporary files
./cleanup.sh
```

## How It Works

### 1. Fetch Zone Data via API

```bash
# Lists all zones to find the identifier
GET /api/clouddns/v1/zone.json?limit=1000

# Fetches all records for the specific zone
GET /api/clouddns/v1/zone.json/<zone_identifier>
```

**Why:** We need the zone identifier and record identifiers to generate correct import blocks. The zone name alone isn't sufficient - Terraform requires the API's internal UUIDs.

### 2. Generate Import Blocks

Creates `_import-blocks.tf` containing:

**Resource Stubs** (placeholder configurations):
```hcl
resource "anxcloud_dns_zone" "zone" {
  name         = "example.com"
  is_master    = true              # Placeholder
  dns_sec_mode = "unvalidated"     # Placeholder
  admin_email  = "admin@example.com" # Placeholder
  refresh      = 3600              # Placeholder
  # ... more placeholders
}
```

**Why placeholders?** Terraform requires resource blocks to exist before importing, but we don't need accurate values. During import, Terraform calls the provider's Read function which fetches the REAL values from the API and replaces these placeholders.

**Import Blocks** (map resources to remote objects):
```hcl
import {
  to = anxcloud_dns_zone.zone
  id = "a94c1c189b924828a80a850d8ee557e0"  # Real zone identifier from API
}

import {
  to = anxcloud_dns_record.www_a_record
  id = "example.com/abc123def456"  # Real record identifier from API
}
```

**Why:** These tell Terraform which remote resources to import and where to store them in state.

### 3. Validate Plan Safety (Auto Mode)

```bash
tofu plan -out=import.tfplan
```

Parses the plan output and validates:
- `0 to add` - No new resources created
- `0 to destroy` - No resources deleted
- `X to import` - Only import operations

**Why:** Safety check to prevent accidental infrastructure changes. If validation fails, auto mode aborts immediately. Saves the plan to a file for apply.

### 4. Import Resources

```bash
tofu apply import.tfplan
```

For each import block, Terraform:
1. Calls the provider's `Read` function
2. Provider queries Anexia Cloud API for the resource
3. Provider returns the real configuration
4. Terraform stores it in `terraform.tfstate`

**Why:** This is how Terraform learns about existing infrastructure. The state file now contains the actual resource configuration. Using the saved plan ensures exactly what was validated gets applied.

### 5. Generate Final Configuration

```bash
tofu plan -generate-config-out=_generated.tf
```

Terraform:
1. Reads resource data from `terraform.tfstate`
2. Generates `.tf` resource blocks with the actual values
3. Writes them to `_generated.tf`

**Why:** You need `.tf` files to continue managing resources. State alone isn't enough - Terraform requires both state AND configuration files.

### 6. Cleanup

Removes temporary files:
- `_import-blocks.tf` - No longer needed after import
- `_temp_main.tf` - Temporary provider config
- `.terraform/` - Provider cache
- `.terraform.lock.hcl` - Lock file

**Why:** Keeps your directory clean. The important files (`imported_*.tf` and `terraform.tfstate`) are preserved.

## Record Filtering

### Excluded Record Types

**SOA (Start of Authority)** and **NS (Name Server)** records are automatically filtered out.

**Why:**
- These are managed automatically by the DNS zone itself
- Cannot be modified independently through the provider
- Importing them would cause Terraform errors on subsequent plans

### TXT Record Quote Handling

The script strips outer quotes from TXT record values:

```
API returns:  "\"my text value\""
Script uses:  my text value
```

**Why:** The provider automatically adds quotes when creating TXT records. If we kept the API's quotes, we'd end up with double-escaped quotes causing drift between state and configuration.

### Resource Naming

Records with duplicate names get numbered suffixes:

```hcl
resource "anxcloud_dns_record" "mail_txt_record"      # First
resource "anxcloud_dns_record" "mail_txt_record_2"    # Second
resource "anxcloud_dns_record" "mail_txt_record_3"    # Third
```

**Why:** Terraform resource names must be unique within a module. Multiple records can have the same DNS name (e.g., multiple MX records), but need unique Terraform identifiers.

## File Reference

### Temporary Files (Removed by Cleanup)

| File | Purpose | When Removed |
|------|---------|--------------|
| `_import-blocks.tf` | Import blocks and resource stubs | After import completes |
| `_temp_main.tf` | Provider config (when run outside example dir) | After import completes |
| `_generated.tf` | Generated config (manual mode only) | **You must rename this!** |
| `.terraform/` | Provider plugin cache | After import completes |
| `.terraform.lock.hcl` | Provider version lock | After import completes |

### Permanent Files (Keep These!)

| File | Purpose | Created By |
|------|---------|------------|
| `imported_dns_zone.tf` | Zone resource configuration | Auto mode |
| `imported_dns_records.tf` | Record configurations | Auto mode |
| `terraform.tfstate` | Resource state | Terraform import |
| Your custom `.tf` files | Resource configurations | Manual mode (you create) |

## Safety Features

### Cleanup Protection

The `cleanup.sh` script warns you if no permanent configuration exists:

```bash
⚠️  WARNING: No permanent Terraform configuration files found!

Before running cleanup, you should:
  1. Rename _generated.tf to a permanent name (e.g., dns_config.tf)
     OR
  2. Create your own .tf files with the imported resource configuration

Are you sure you want to continue? (y/N)
```

**Why:** Prevents accidentally removing your only configuration, which would leave resources in state but no way to manage them.

### Auto Mode Validation

Auto mode validates the plan before importing:

```
Plan analysis:
  - To add: 0
  - To change: 0
  - To destroy: 0
  - To import: 6

✅ SAFE: Plan contains only import operations
```

If the plan shows any `add` or `destroy` operations, auto mode aborts.

**Why:** Ensures the import is truly read-only and won't modify your infrastructure.

## Troubleshooting

### "Zone already exists" Error

**Status:** Fixed in latest provider version.

The provider now allows importing existing zones. If you still see this error, rebuild the provider:

```bash
cd /path/to/terraform-provider-anxcloud
make install
```

### No Configuration After Cleanup

**Symptoms:** `tofu plan` fails with "no configuration files found"

**Cause:** You ran `cleanup.sh` before saving `_generated.tf`

**Recovery:**
```bash
# Resources are still in state
tofu state list

# Create minimal provider config
cat > provider.tf << 'EOF'
terraform {
  required_providers {
    anxcloud = {
      source  = "anexia-it/anxcloud"
      version = "99.0.0"
    }
  }
}

provider "anxcloud" {}
EOF

# Re-initialize
tofu init

# Manually create resource blocks based on state
tofu state show anxcloud_dns_zone.zone
```

### Import Shows "Will Be Created"

**Symptoms:** Plan shows resources will be created instead of imported

**Cause:** Import blocks are missing from `_import-blocks.tf`

**Fix:** Regenerate the import blocks:
```bash
./cleanup.sh
./generate-import-blocks.sh example.com
```

### Authentication Errors

Ensure your token is set:
```bash
export ANEXIA_TOKEN='your-token-here'
export ANEXIA_BASE_URL='https://engine.anexia-it.com'  # or your environment URL
```

For integration environments, use `Authorization: Token` (not `Bearer`).

## Advanced Usage

### Import from Anywhere

```bash
cd ~/my-terraform-project
/path/to/generate-import-blocks.sh --auto example.com
```

The script creates files in your current directory.

### Import Multiple Zones

```bash
./generate-import-blocks.sh --auto zone1.com
./generate-import-blocks.sh --auto zone2.com
./generate-import-blocks.sh --auto zone3.com
```

Each zone gets its own `imported_dns_zone_<zonename>.tf` and `imported_dns_records_<zonename>.tf` files (in auto mode).

### Custom Provider Version

Edit `generate-import-blocks.sh` line 123:
```bash
version = "99.0.0"  # Change to your desired version
```

### Integration with Remote State

The script uses local state (`-backend=false`). To use remote state:

1. Run the import with local state
2. Copy the generated `.tf` files to your main project
3. Import into your remote state:
```bash
terraform import anxcloud_dns_zone.zone <zone_identifier>
terraform import anxcloud_dns_record.www_a_record example.com/<record_identifier>
```

## Examples

### Basic Import

```bash
./generate-import-blocks.sh --auto example.com
# Output: imported_dns_zone.tf, imported_dns_records.tf
tofu plan  # Should show "No changes"
```

### Manual Review Before Import

```bash
./generate-import-blocks.sh example.com
cat _import-blocks.tf  # Review import blocks
tofu plan -out=import.tfplan -generate-config-out=_generated.tf
cat _generated.tf      # Review generated config
tofu apply import.tfplan
mv _generated.tf dns_config.tf
./cleanup.sh
```

### Import into Existing Project

```bash
cd ~/existing-terraform-project
/path/to/generate-import-blocks.sh --auto myzone.com
# Files created in current directory
git add imported_dns_*.tf terraform.tfstate
git commit -m "Import DNS zone myzone.com"
```

## Technical Details

### Import ID Format

**Zone:** `<zone_identifier>`
```
Example: a94c1c189b924828a80a850d8ee557e0
```

**Record:** `<zone_name>/<record_identifier>`
```
Example: example.com/abc123def456789
```

**Why this format?** The provider's `Importer.StateContext` function expects this format to parse the zone name and record ID.

### API Pagination

The script fetches up to 1000 zones with `?limit=1000`. If you have more zones, you must specify the zone name directly.

### Provider Version

Default: `99.0.0` (local development version)

Adjust in `_temp_main.tf` or use a `.terraform.lock.hcl` file for production versions.

### State Backend

The script uses `-backend=false` to avoid remote state configuration. Integrate with your backend as needed.
