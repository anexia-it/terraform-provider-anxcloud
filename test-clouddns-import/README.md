# CloudDNS Import Testing

This directory contains a simple Terraform configuration to test the CloudDNS record import functionality implemented in the `feature/clouddns-stable-identifiers` branch.

## Prerequisites

1. **AnxCloud API Access**: You need access to the AnxCloud integration environment
2. **API Token**: Set the `ANEXIA_TOKEN` environment variable with your API token
3. **Terraform or OpenTofu**: Install either Terraform (version 1.x) or OpenTofu (the script auto-detects which one you have)
4. **Local Provider Build**: Build and install the local provider from the `feature/clouddns-stable-identifiers` branch

## Setup

1. **Build and install the local provider** (from the main project directory):
   ```bash
   git checkout feature/clouddns-stable-identifiers
   make build
   make install
   ```

2. **Set environment variables**:
   ```bash
   export ANEXIA_TOKEN="your-api-token-here"
   export ANEXIA_BASE_URL="https://integration-1.anexia-it.com"  # or your test environment
   ```

3. **Navigate to test directory**:
   ```bash
   cd test-clouddns-import
   ```

4. **Configure Terraform/OpenTofu to use local provider**:
   The test directory includes a `dev.tfrc` file that configures development overrides to bypass checksum verification. This allows you to rebuild the provider without lock file conflicts.
   provider_installation {
     dev_overrides {
       "anexia-it/anxcloud" = "/path/to/terraform-provider-anxcloud"
     }
     direct {}
   }
   ```

## Test Scenario: Create and Import

### Step 1: Create Resources
```bash
terraform init
terraform plan
terraform apply
```

This will create:
- A DNS zone: `test-import-zone.terraform.example`
- A DNS record: `test-record.test-import-zone.terraform.example` → `192.168.1.100`

### Step 2: Verify Resources
Check that the resources were created and note the stable identifier:
```bash
terraform show
```

Look for the `identifier` field in the DNS record output - this is the stable UUID you can use for importing.

### Step 3: Test Import (Option A - Import Existing Resource)
Remove the resource from Terraform state but keep it in the API:
```bash
# Remove from state only
terraform state rm anxcloud_dns_record.test_record

# Now import it back using the stable identifier
terraform import anxcloud_dns_record.test_record <stable-identifier>

# Verify it imported correctly
terraform plan  # Should show no changes
```

### Step 4: Test Import (Option B - Import External Resource)
If you have an existing DNS record in your AnxCloud environment:

1. Create a new Terraform resource block for an existing record:
   ```hcl
   resource "anxcloud_dns_record" "existing_record" {
     name      = "existing-record-name"
     zone_name = "existing-zone-name"
     type      = "A"
     rdata     = "1.2.3.4"  # Must match existing record
     ttl       = 300       # Must match existing record
   }
   ```

2. Import using the new format:
   ```bash
   terraform import anxcloud_dns_record.existing_record <zone_name>/<stable-identifier>
   ```

   For example:
   ```bash
   terraform import anxcloud_dns_record.existing_record example.com/abc123def-4567-8901-2345-6789abcdef01
   ```

## Verification

After import, run:
```bash
terraform plan
```

The output should show no changes, confirming that:
1. The import worked correctly
2. The stable identifier is being used
3. The resource state matches the API state

## Cleanup

Remove the test resources:
```bash
terraform destroy
```

## Troubleshooting

### Common Issues

1. **"Resource already managed"**: If you try to import a resource that's already in state
2. **"Record not found"**: Check that the stable identifier is correct
3. **"Zone not found"**: Ensure the zone exists and the zone_name matches exactly

### Finding Stable Identifiers

You can find stable identifiers by:
1. Looking at `terraform show` output after creating resources
2. Using the AnxCloud API directly
3. Checking existing Terraform state files

### Debug Mode

Enable debug logging:
```bash
export TF_LOG=DEBUG
terraform import anxcloud_dns_record.test_record <identifier>
```

## Expected Behavior

With the stable identifiers implementation:
- ✅ Import should work using stable UUIDs
- ✅ Identifiers should persist across record updates
- ✅ Performance should be better (identifier-based lookups)
- ✅ Backward compatibility should be maintained

## Files

- `main.tf`: Terraform configuration for testing
- `README.md`: This documentation</content>
<parameter name="filePath">test-clouddns-import/README.md