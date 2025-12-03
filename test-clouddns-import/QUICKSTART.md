# Quick Start Guide - CloudDNS Import Testing

## 🚀 Quick Test (Automated)

```bash
cd test-clouddns-import
export ANEXIA_TOKEN="your-token-here"
export ANEXIA_BASE_URL="https://integration-1.anexia-it.com"
./test-import.sh full-test
```

## 📋 Step-by-Step Manual Testing

### 1. Build Local Provider
```bash
cd /home/rweselowski/PhpstormProjects/terraform-provider-anxcloud
git checkout feature/clouddns-stable-identifiers
make install
```

### 2. Setup Environment
```bash
cd test-clouddns-import
export ANEXIA_TOKEN="your-token-here"
export ANEXIA_BASE_URL="https://integration-1.anexia-it.com"
```

### 3. Run Tests
```bash
# Build provider
./test-import.sh build

# Initialize Terraform
./test-import.sh init

# Create resources
./test-import.sh create

# Test import functionality
./test-import.sh test-import

# Cleanup
./test-import.sh cleanup
```

## 🔍 What Gets Tested

1. **Provider Build**: Local provider from feature branch
2. **Resource Creation**: DNS zone + DNS record
3. **Stable Identifier**: UUID identifier is populated
4. **Import Functionality**: Remove from state, import back using identifier
5. **State Consistency**: Terraform plan shows no changes after import

## ✅ Expected Results

After `test-import`:
- Import succeeds without errors
- `terraform plan` shows: **No changes. Your infrastructure matches the configuration.**
- Stable identifier persists in state

## 📊 Verification

Check the outputs:
```bash
terraform output dns_record_identifier  # Shows stable UUID
terraform output dns_record_id          # Shows Terraform ID
```

## 🛠️ Provider Configuration

The test uses:
- **Source**: `hashicorp.com/anexia-it/anxcloud`
- **Version**: `0.3.1`
- **Location**: `~/.terraform.d/plugins/hashicorp.com/anexia-it/anxcloud/0.3.1/linux_amd64/`

This matches the Makefile configuration for local development.