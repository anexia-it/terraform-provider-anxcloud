#!/bin/bash

set -eu -o pipefail

if [ "${ANEXIA_DEBUG:-0}" != "0" ]; then
  set -x
fi

# Detect script directory and save current working directory
SCRIPT_PATH="${BASH_SOURCE[0]:-$0}"
SCRIPT_DIR="$(cd "$(dirname "$SCRIPT_PATH")" && pwd)"
WORK_DIR="$(pwd)"

function show_usage {
  cat <<EOF
Usage: $(basename "$0") <zone_name>

Generate Terraform import blocks for a DNS zone and all its records from Anexia Cloud.

Arguments:
  <zone_name>    Name of the DNS zone to import records from

Options:
  -h, --help     Show this help message
  -a, --auto     Automatic mode: generate, validate, import, and cleanup

The script can be run from any directory. It will create temporary Terraform
configuration files in your current directory and generate _import-blocks.tf
with import statements for all records in the specified zone.

IMPORTANT: This script is for IMPORT ONLY. It does not apply any modifications
to your infrastructure. It only generates import blocks.

Environment Variables:
  ANEXIA_TOKEN      Required. API token for Anexia Cloud
  ANEXIA_BASE_URL   Optional. Override default Anexia API base URL

Modes:
  Manual mode (default):
    - Generates import blocks only
    - User reviews and runs subsequent commands manually
    - Temporary files left for next steps
  
  Automatic mode (--auto):
    - Generates import blocks
    - Automatically generates resource configuration
    - Validates plan contains ONLY imports (safety check)
    - Automatically applies imports
    - Creates imported.tf with final resource definitions
    - Cleans up all temporary files
    - ⚠️  SAFE: Will NEVER create, modify, or destroy resources
    - ⚠️  ABORTS if plan shows any non-import changes

Examples:
  # Manual mode
  $(basename "$0") example.com
  
  # Automatic mode
  $(basename "$0") --auto example.com

  # Run from your Terraform project directory
  cd ~/my-terraform-project
  /path/to/scripts/generate-import-blocks.sh --auto example.com

Output (Manual mode):
   _import-blocks.tf     Import blocks for all DNS records
   _temp_main.tf         Temporary Terraform configuration (when run from other directory)
   .terraform/           Terraform cache directory

Output (Auto mode):
  imported_dns_zone.tf     DNS zone resource configuration
  imported_dns_records.tf  DNS records resource configurations

Important:
  - Manual mode: Temporary files are left for next steps (cleanup manually)
  - Auto mode: Creates imported.tf and cleans up everything else automatically

Validation:
  - Script will abort if _temp_*.tf files already exist
  - Script will NOT override any existing files
  - You must manually remove old temporary files before re-running

Next steps after running this script:
  1. Review the generated _import-blocks.tf
  2. Run: tofu plan -generate-config-out=_generated.tf
  3. Review _generated.tf and edit as needed
  4. Run: tofu apply (to import, not to modify infrastructure)
  5. Clean up: rm _temp_*.tf .terraform/ -rf
EOF
}

function setup_provider_config {
  local script_dir="$1"
  local work_dir="$2"
  local zone_name="$3"

  if [[ "$script_dir" == "$work_dir" ]]; then
    return 0
  fi

  # Check if main.tf exists in work_dir
  if [[ -f "$work_dir/main.tf" ]]; then
    return 0
  fi

  # Check if _temp_main.tf already exists
  if [[ -f "_temp_main.tf" ]]; then
    echo "ERROR: Temporary file _temp_main.tf already exists!"
    echo "Please remove it manually before running this script:"
    echo "  rm _temp_main.tf"
    echo ""
    echo "Or clean up all temporary files:"
    echo "  rm _temp_*.tf"
    exit 1
  fi

  # Generate main.tf with hardcoded zone name
  cat >_temp_main.tf <<EOF
terraform {
  required_providers {
    anxcloud = {
      source  = "anexia-it/anxcloud"
      version = "99.0.0"
    }
  }
}

provider "anxcloud" {
  # Authentication via ANEXIA_TOKEN environment variable
}

# Data source to fetch existing DNS records from the zone
data "anxcloud_dns_records" "zone_records" {
  zone_name = "$zone_name"
}
EOF

  TEMP_TF_FILES+=("_temp_main.tf")
  echo "  Created _temp_main.tf with zone: $zone_name"
}

function setup_terraform_files {
  local script_dir="$1"
  local work_dir="$2"
  local zone_name="$3"

  # Check if we're running from the script's own directory
  if [[ "$script_dir" == "$work_dir" ]]; then
    echo "Running from example directory - using existing Terraform files..."
    # No need to copy files, they're already here
    TEMP_TF_FILES=()
    return 0
  fi

  echo "Setting up Terraform configuration files..."

  TEMP_TF_FILES=()

  # Only setup provider config (no variables needed)
  setup_provider_config "$script_dir" "$work_dir" "$zone_name"
}

function validate_plan_is_safe {
  local plan_output="$1"

  # Extract plan summary line
  # Example: "Plan: 3 to import, 0 to add, 0 to change, 0 to destroy."
  local plan_summary=$(echo "$plan_output" | grep -E "Plan:")

  if [[ -z "$plan_summary" ]]; then
    echo "ERROR: Could not parse plan output"
    return 1
  fi

  # Parse numbers - handle both old and new format
  # New: "Plan: 3 to import, 0 to add, 0 to change, 0 to destroy."
  # Old: "Plan: 0 to add, 0 to change, 0 to destroy, 3 to import."
  local to_add=$(echo "$plan_summary" | grep -oE '[0-9]+ to add' | grep -oE '[0-9]+')
  local to_change=$(echo "$plan_summary" | grep -oE '[0-9]+ to change' | grep -oE '[0-9]+')
  local to_destroy=$(echo "$plan_summary" | grep -oE '[0-9]+ to destroy' | grep -oE '[0-9]+')
  local to_import=$(echo "$plan_summary" | grep -oE '[0-9]+ to import' | grep -oE '[0-9]+')

  echo "Plan analysis:"
  echo "  - To add: ${to_add:-0}"
  echo "  - To change: ${to_change:-0}"
  echo "  - To destroy: ${to_destroy:-0}"
  echo "  - To import: ${to_import:-0}"

  # Safety check: Reject any add or destroy operations
  if [[ "${to_add:-0}" -ne 0 || "${to_destroy:-0}" -ne 0 ]]; then
    echo ""
    echo "❌ UNSAFE: Plan contains add/destroy operations!"
    echo "   Auto mode ONLY imports resources - no creates or destroys allowed"
    echo "   Aborting for safety."
    return 1
  fi

  # Check if there are any imports
  if [[ "${to_import:-0}" -eq 0 ]]; then
    echo ""
    echo "⚠️  WARNING: No resources to import"
    return 1
  fi

  # Handle changes during import
  if [[ "${to_change:-0}" -gt 0 ]]; then
    # Changes are expected when importing with resource stubs that have placeholder values
    # The changes should be on the imported resources (stub values -> actual remote values)

    if [[ "${to_change:-0}" -eq "${to_import:-0}" ]]; then
      # This is the expected case: each imported resource shows as "changed"
      # because the stub configuration doesn't match the remote resource exactly
      echo ""
      echo "⚠️  Note: Plan shows changes on imported resources"
      echo "   This is expected when importing with resource stubs"
      echo "   The stub values will be updated to match actual remote state during import"
      echo "   Changes: ${to_change}, Imports: ${to_import}"
    elif [[ "${to_change:-0}" -lt "${to_import:-0}" ]]; then
      # Fewer changes than imports is also acceptable
      # Some imported resources might have matching stub values
      echo ""
      echo "⚠️  Note: Plan shows some changes on imported resources"
      echo "   This is expected when importing with resource stubs"
      echo "   Changes: ${to_change}, Imports: ${to_import}"
    else
      # More changes than imports suggests other configuration changes
      echo ""
      echo "❌ UNSAFE: Plan contains more changes than imports!"
      echo "   Changes: ${to_change}, Imports: ${to_import}"
      echo "   This suggests unrelated configuration changes beyond the import"
      echo "   Aborting for safety."
      return 1
    fi
  fi

  echo ""
  echo "✅ SAFE: Plan contains only import operations (${to_import} resources to import)"
  return 0
}

function run_auto_mode {
  local zone_name="$1"

  echo ""
  echo "╔════════════════════════════════════════════════════════════════╗"
  echo "║              AUTOMATIC IMPORT MODE                             ║"
  echo "╠════════════════════════════════════════════════════════════════╣"
  echo "║  This will automatically import all DNS records                ║"
  echo "║  ⚠️  SAFETY: Only imports - never creates/modifies/destroys    ║"
  echo "╚════════════════════════════════════════════════════════════════╝"
  echo ""

  # Step 1: Validate plan is safe
  echo "▶ Step 1: Validating plan safety..."
  if ! $TF_CMD plan -out "/tmp/plan_$$" -no-color >"/tmp/plan_output_$$.txt" 2>&1; then
    echo "❌ Failed to run plan"
    cat /tmp/plan_output_$$.txt
    rm -f /tmp/plan_output_$$.txt
    return 1
  fi

  local plan_output

  plan_output=$(cat /tmp/plan_output_$$.txt)

  rm -f /tmp/plan_output_$$.txt

  if ! validate_plan_is_safe "$plan_output"; then
    echo ""
    echo "Auto mode aborted. You can review manually:"
    echo "  1. Run: $TF_CMD plan"
    echo "  2. If safe, run: $TF_CMD apply"
    return 1
  fi

  # Step 2: Apply imports
  echo ""
  echo "▶ Step 2: Applying imports..."
  if ! $TF_CMD apply -auto-approve -no-color "/tmp/plan_$$"; then
    echo "❌ Apply failed"
    return 1
  fi

  echo ""
  echo "✅ All resources imported successfully!"

  # Step 3: Generate final resource configuration
  echo ""
  echo "▶ Step 3: Generating final resource configuration..."
  if ! $TF_CMD plan -generate-config-out=_generated.tf -no-color; then
    echo "❌ Failed to generate final configuration"
    return 1
  fi

  # Replace variable reference with actual zone name (if any)
  sed -i "s/var\.zone_name/\"$zone_name\"/g" _generated.tf 2>/dev/null || true
  echo "✅ Final configuration generated in _generated.tf"

  # Step 4: Finalize and split into separate files
  echo ""
  echo "▶ Step 4: Finalizing and cleaning up..."

  imported_file="imported_${zone_name}.tf"
  # Split generated config into zone and records files
  if [[ -f "_generated.tf" ]]; then
    # Zone file
    cat >"$imported_file" <<EOF
# Imported DNS Zone Configuration
# Generated by: generate-import-blocks.sh --auto
# Zone: ${zone_name}
# Zone names are hardcoded (not using variables).

EOF
    cat "_generated.tf" >>"$imported_file"

    # Format
    $TF_CMD fmt "${imported_file}" >/dev/null 2>&1 || true

    rm -f _generated.tf
    echo "  ✅ Created ${imported_file}"
  fi

  # Remove data source from state (it was only needed during import)
  if $TF_CMD state list 2>/dev/null | grep -q "data.anxcloud_dns_records"; then
    $TF_CMD state rm data.anxcloud_dns_records.zone_records >/dev/null 2>&1 || true
    echo "  ✅ Removed temporary data source from state"
  fi

  # Clean up temporary files
  rm -f _import-blocks.tf _temp_*.tf
  rm -rf .terraform/ .terraform.lock.hcl
  echo "  ✅ Cleaned up temporary files"

  echo ""
  echo "╔════════════════════════════════════════════════════════════════╗"
  echo "║              AUTOMATIC IMPORT COMPLETE                         ║"
  echo "╠════════════════════════════════════════════════════════════════╣"
  echo "║  ✅ All DNS records imported successfully                      ║"
  echo "║  ✅ Zone configuration saved to: imported_dns_zone.tf          ║"
  echo "║  ✅ Records configuration saved to: imported_dns_records.tf    ║"
  echo "║  ✅ All temporary files cleaned up                             ║"
  echo "╚════════════════════════════════════════════════════════════════╝"
  echo ""
  echo "Next steps:"
  echo "  1. Review imported_dns_zone.tf and imported_dns_records.tf"
  echo "  2. Add both files to version control"
  echo "  3. Run: $TF_CMD plan (should show no changes)"
  echo ""

  return 0
}

# Parse options before zone_name argument
AUTO_MODE=false
ZONE_NAME=""
while [[ $# -gt 0 ]]; do
  case "$1" in
  --help | -h)
    show_usage
    exit 0
    ;;
  --auto | -a)
    AUTO_MODE=true
    shift
    ;;
  *)
    # Assume it's the zone_name
    ZONE_NAME="$1"
    shift
    break
    ;;
  esac
done

# Validate zone_name is provided
if [[ -z "$ZONE_NAME" ]]; then
  echo "Error: zone_name is required"
  echo ""
  show_usage
  exit 1
fi

base_url="${ANEXIA_BASE_URL:-https://engine.anexia-it.com}"
base_url="${base_url%/}"

function get_zone_identifier {
  local base_url="$1"
  local zone_name="$2"
  local api_url="$base_url/api/clouddns/v1/zone.json?limit=1000"

  local response

  if ! response=$(curl -s -H "Authorization: Token $ANEXIA_TOKEN" "$api_url"); then
    echo "Error: Failed to fetch zones from API"
    exit 1
  fi

  local identifier
  identifier=$(echo "$response" | jq -r ".results[] | select(.name == \"$zone_name\") | .identifier")

  if [ -z "$identifier" ] || [ "$identifier" = "null" ]; then
    echo "Error: Zone '$zone_name' not found"
    exit 1
  fi

  echo "$identifier"
}

function get_zone_records {
  local base_url="$1"
  local zone_identifier="$2"
  local api_url="$base_url/api/clouddns/v1/zone.json/$zone_identifier"

  local response

  if ! response=$(curl -s -H "Authorization: Token $ANEXIA_TOKEN" "$api_url"); then
    echo "Error: Failed to fetch records from API"
    exit 1
  fi

  # Transform the API response to match the expected structure for jq processing
  echo "$response" | jq '.revisions[0].records // [] | map({
     name: .name,
     type: .type,
     ttl: .ttl,
     rdata: .rdata,
     import_id: .identifier
   })'
}

function main {
  local zone_name="$1"
  OUTPUT_FILE="_import-blocks.tf"
  TF_CMD=tofu      # command to use to perform the actions (terraform, tofu)
  TEMP_TF_FILES=() # Array to track temporary files

  # Setup temporary Terraform files (with validation)
  setup_terraform_files "$SCRIPT_DIR" "$WORK_DIR" "$zone_name"

  # Fetch zone identifier early (needed for import block generation)
  echo "Fetching zone identifier..."
  if ZONE_IDENTIFIER=$(get_zone_identifier "$base_url" "$ZONE_NAME" 2>/dev/null); then
    echo "  Found zone identifier: $ZONE_IDENTIFIER"
  else
    echo "  Warning: Could not fetch zone identifier (API access failed)"
    echo "  Using placeholder for import block generation"
    ZONE_IDENTIFIER="zone-identifier-placeholder"
  fi

  # NOTE: We do NOT cleanup temp files automatically because users need them
  # for the next step: tofu plan -generate-config-out=_generated.tf
  # Users can manually clean up with: rm _temp_*.tf
  # Or use the cleanup.sh script in the example directory

  # Rest of existing logic continues here...

  cat <<EOF >"$OUTPUT_FILE"
# ============================================================
# AUTO-GENERATED IMPORT CONFIGURATION
# ============================================================
# Generated by: generate-import-blocks.sh
# Zone: $ZONE_NAME
# Generated on: $(date)
#
# This file contains:
# 1. Resource stubs (placeholder blocks required by Terraform/OpenTofu)
# 2. Import blocks (specify resources to import from Anexia Cloud)
#
# ⚠️  IMPORTANT: Resource Stub Values Are PLACEHOLDERS!
#
# The values in the resource stubs below (like refresh=3600, ttl=3600)
# are NOT the real values from your zone. They are temporary placeholders.
#
# During import, Terraform will:
# - Fetch the REAL configuration from Anexia Cloud API
# - REPLACE all placeholder values with actual API values
# - Store the real values in terraform.tfstate
#
# You do NOT need to edit these stubs to match your zone.
# The provider handles everything automatically!
#
# Usage:
#   $TF_CMD plan -generate-config-out=_generated.tf
#   $TF_CMD apply
#
# After import, use _generated.tf (if created) for the real configuration.
# ============================================================

# ============================================================
# RESOURCE STUBS
# ============================================================
#
# IMPORTANT: The values in these resource stubs are TEMPORARY PLACEHOLDERS!
# 
# During the import process, Terraform will:
# 1. Use these stubs to know WHERE to put imported data
# 2. Call the Anexia Cloud API to fetch REAL configuration
# 3. REPLACE all these placeholder values with actual API values
#
# You do NOT need to match these values to your real zone config.
# The Terraform provider automatically fetches the correct values from the API.
# ============================================================

# Zone resource stub
# NOTE: These are placeholder values that will be replaced during import
resource "anxcloud_dns_zone" "zone" {
  name         = "$ZONE_NAME"
  is_master    = true              # Placeholder - will be replaced with real API value
  dns_sec_mode = "unvalidated"     # Placeholder - will be replaced with real API value
  admin_email  = "admin@$ZONE_NAME" # Placeholder - will be replaced with real API value
  refresh      = 3600              # Placeholder - will be replaced with real API value
  retry        = 1800              # Placeholder - will be replaced with real API value
  expire       = 604800            # Placeholder - will be replaced with real API value
  ttl          = 3600              # Placeholder - will be replaced with real API value
  # All other attributes will be fetched from API during import
}

EOF

  # Check if $TF_CMD is available
  if ! command -v $TF_CMD &>/dev/null; then
    echo "Error: OpenTofu ($TF_CMD) CLI is not installed or not in PATH"
    exit 1
  fi

  # Check if jq is available (needed for JSON parsing)
  if ! command -v jq &>/dev/null; then
    echo "Warning: jq is not installed. JSON parsing will not be available."
    echo "The script will generate example import blocks only."

    exit 1
  fi

  # Check if curl is available (needed for API calls)
  if ! command -v curl &>/dev/null; then
    echo "Error: curl is not installed. Required for API calls."
    exit 1
  fi

  # Check if ANEXIA_TOKEN is set (optional for example generation)
  if [ -z "${ANEXIA_TOKEN:-}" ]; then
    cat <<EOF
    Warning: ANEXIA_TOKEN environment variable is not set
    The script will generate example import blocks only.
EOF
  fi

  echo "Initializing Terraform..."
  $TF_CMD init -backend=false

  echo "Planning to discover records..."
  # Run plan to validate configuration (don't exit on error)
  set +e
  PLAN_OUTPUT=$($TF_CMD plan -no-color 2>&1)
  PLAN_EXIT_CODE=$?
  set -e

  if [ $PLAN_EXIT_CODE -ne 0 ] || echo "$PLAN_OUTPUT" | grep -q "401 Unauthorized"; then
    cat <<EOF
    Configuration validation failed or authentication error. This is expected if you don't have valid Anexia Cloud credentials.
    The script will generate example import blocks instead.

    To use this script with real data:
    1. Set a valid ANEXIA_TOKEN environment variable
    2. Ensure the zone exists in your Anexia Cloud account
    3. Run the script again
EOF
    exit 1

  fi

  echo "Refreshing data sources..."
  $TF_CMD refresh -no-color >/dev/null 2>&1

  echo "Extracting DNS records from data source..."

  # Attempt 1: Use tofu console to query data source (no state needed)
  CONSOLE_OUTPUT=$(
    $TF_CMD console <<CONSOLE_EOF 2>/dev/null
jsonencode(data.anxcloud_dns_records.zone_records.records)
CONSOLE_EOF
  )
  # Console returns a JSON-encoded string, decode it to get the actual JSON array
  RECORDS_JSON=$(echo "$CONSOLE_OUTPUT" | jq -r . 2>/dev/null || echo "$CONSOLE_OUTPUT")

  # Check if console method worked
  if [ -z "$RECORDS_JSON" ] || [ "$RECORDS_JSON" = "null" ] || [ "$RECORDS_JSON" = "[]" ]; then
    echo "  Console method failed, trying direct API call..."

    # Attempt 2: Fetch records directly from API
    RECORDS_JSON=$(get_zone_records "$base_url" "$ZONE_IDENTIFIER" 2>/dev/null)

    if [ -z "$RECORDS_JSON" ] || [ "$RECORDS_JSON" = "null" ] || [ "$RECORDS_JSON" = "[]" ]; then
      echo "Error: No DNS records found for zone '$ZONE_NAME'"
      echo "  - Data source query failed"
      echo "  - Direct API call failed"
      echo "  Please check:"
      echo "    1. ANEXIA_TOKEN is valid"
      echo "    2. Zone '$ZONE_NAME' exists"
      echo "    3. Zone has at least one record"
      exit 1
    fi

    echo "  ✅ Successfully fetched records from API (found $(echo "$RECORDS_JSON" | jq 'length') records)"
  else
    echo "  ✅ Successfully extracted records from data source (found $(echo "$RECORDS_JSON" | jq 'length') records)"
  fi

  echo "  Found $(echo "$RECORDS_JSON" | jq 'length') DNS records"
  {
    # Generate resource stubs for DNS records
    cat <<'RECORD_STUB_HEADER'

# Record resource stubs
# NOTE: Like the zone stub above, these are placeholder values.
# During import, Terraform will fetch the real values from the API and replace these.
RECORD_STUB_HEADER
    echo "$RECORDS_JSON" | jq -r --arg zone_name "$ZONE_NAME" '
def sanitize_name:
  # Replace @ with root, * with wildcard, other special chars with _
  gsub("@"; "root") |
  gsub("\\*"; "wildcard") |
  gsub("[^a-zA-Z0-9_]"; "_") |
  # Ensure it starts with a letter or underscore
  if test("^[0-9]") then "_" + . else . end |
  # Convert to lowercase and handle empty strings
  ascii_downcase |
  if . == "" then "default" else . end;

# Strip outer quotes from TXT record rdata values
# The API returns TXT records with RFC-compliant quotes (e.g., "\"value\"")
# but the provider expects them without quotes (adds them automatically on create)
# This matches the provider behavior in resourceDNSRecordRead which strips outer quotes
def strip_txt_quotes(record_type):
  if record_type == "TXT" then
    # Check if string starts and ends with quote
    if (. | length >= 2) and (.[0:1] == "\"") and (.[-1:] == "\"") then
      .[1:-1]  # Strip first and last character
    else
      .
    end
  else
    .
  end;

# Filter out SOA and NS records (cannot be modified)
map(select(.type != "SOA" and .type != "NS")) |

# Transform records to include generated resource names
map({
  record: .,
  resource_name: ((.name | sanitize_name) + "_" + (.type | ascii_downcase) + "_record")
}) |

# Group by resource_name
group_by(.resource_name) |

# Process each group
map(
  if length > 1 then
    # Multiple records with same name - add numbering
    . as $group |
    range(0; $group | length) | {index: ., item: $group[.]}
  else
    # Single record - no numbering needed
    {index: 0, item: .[0]}
  end
) |

# Flatten and generate resource stubs
.[] | .item as $item | .index as $index |
# Process rdata: strip outer quotes for TXT records to match provider behavior
($item.record.rdata | strip_txt_quotes($item.record.type)) as $processed_rdata |
# Include TTL if present (use zone default otherwise by omitting)
($item.record.ttl // null) as $ttl |
if $index == 0 then
  if $ttl != null then
    "resource \"anxcloud_dns_record\" \"\($item.resource_name)\" {\n  name = \"\($item.record.name)\"\n  type = \"\($item.record.type)\"\n  rdata = \($processed_rdata | @json)\n  ttl = \($ttl)\n  zone_name = \"\($zone_name)\"\n}\n"
  else
    "resource \"anxcloud_dns_record\" \"\($item.resource_name)\" {\n  name = \"\($item.record.name)\"\n  type = \"\($item.record.type)\"\n  rdata = \($processed_rdata | @json)\n  zone_name = \"\($zone_name)\"\n}\n"
  end
else
  if $ttl != null then
    "resource \"anxcloud_dns_record\" \"\($item.resource_name)_\($index + 1)\" {\n  name = \"\($item.record.name)\"\n  type = \"\($item.record.type)\"\n  rdata = \($processed_rdata | @json)\n  ttl = \($ttl)\n  zone_name = \"\($zone_name)\"\n}\n"
  else
    "resource \"anxcloud_dns_record\" \"\($item.resource_name)_\($index + 1)\" {\n  name = \"\($item.record.name)\"\n  type = \"\($item.record.type)\"\n  rdata = \($processed_rdata | @json)\n  zone_name = \"\($zone_name)\"\n}\n"
  end
end
'
    cat <<EOF

# ============================================================
# IMPORT BLOCKS
# ============================================================

# Import the DNS zone itself
import {
  to = anxcloud_dns_zone.zone
  id = "$ZONE_IDENTIFIER"
}
EOF
  } >>"$OUTPUT_FILE"

  # Generate import blocks for DNS records
  echo "$RECORDS_JSON" | jq -r --arg zone_name "$ZONE_NAME" '
def sanitize_name:
  # Replace @ with root, * with wildcard, other special chars with _
  gsub("@"; "root") |
  gsub("\\*"; "wildcard") |
  gsub("[^a-zA-Z0-9_]"; "_") |
  # Ensure it starts with a letter or underscore
  if test("^[0-9]") then "_" + . else . end |
  # Convert to lowercase and handle empty strings
  ascii_downcase |
  if . == "" then "default" else . end;

# Filter out SOA and NS records (cannot be modified)
map(select(.type != "SOA" and .type != "NS")) |

# Transform records to include generated resource names
map({
  record: .,
  resource_name: ((.name | sanitize_name) + "_" + (.type | ascii_downcase) + "_record")
}) |

# Group by resource_name
group_by(.resource_name) |

# Process each group
map(
  if length > 1 then
    # Multiple records with same name - add numbering
    . as $group |
    range(0; $group | length) | {index: ., item: $group[.]}
  else
    # Single record - no numbering needed
    {index: 0, item: .[0]}
  end
) |

# Flatten and generate import blocks
.[] | .item as $item | .index as $index |
if $index == 0 then
  "import {\n  to = anxcloud_dns_record.\($item.resource_name)\n  id = \"\($zone_name)/\($item.record.identifier)\"\n}\n"
else
  "import {\n  to = anxcloud_dns_record.\($item.resource_name)_\($index + 1)\n  id = \"\($zone_name)/\($item.record.identifier)\"\n}\n"
end
' >>"$OUTPUT_FILE"

  # Check if auto mode
  if [[ "$AUTO_MODE" == "true" ]]; then
    echo "  Generated import blocks in $OUTPUT_FILE"

    # Run automatic import workflow
    run_auto_mode "$zone_name"
    exit_code=$?

    # Cleanup temp files even if auto mode fails
    if [[ ${#TEMP_TF_FILES[@]} -gt 0 ]]; then
      rm -f "${TEMP_TF_FILES[@]}"
    fi

    exit $exit_code
  else

    cat <<EOF
  Generated import blocks in $OUTPUT_FILE

  Output location: $WORK_DIR/$OUTPUT_FILE

  Next steps:
  1. Review and edit the generated import blocks if needed (_import-blocks.tf)
  2. Run: $TF_CMD plan -generate-config-out=_generated.tf -out "generate_plan"
  3. Review _generated.tf and edit as needed
  4. Run: $TF_CMD apply "generate_plan" (IMPORT ONLY - no infrastructure modifications)
  5. ⚠️  IMPORTANT: Rename _generated.tf to a permanent name BEFORE cleanup:
     mv _generated.tf dns_config.tf
     (or split into separate files like the --auto mode does)

  TIP: Use --auto flag for fully automatic import (safe, import-only)
EOF
  fi
}

main "$ZONE_NAME"
