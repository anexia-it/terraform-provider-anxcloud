#!/bin/bash

# CloudDNS Import Test Helper Script
# This script helps test the CloudDNS import functionality

set -e

echo "🔍 CloudDNS Import Test Helper"
echo "================================"

# Detect which tool to use (terraform or tofu)
if command -v tofu &> /dev/null; then
    TF_CMD="tofu"
    echo "🔧 Using OpenTofu"
elif command -v terraform &> /dev/null; then
    TF_CMD="terraform"
    echo "🔧 Using Terraform"
else
    echo "❌ Error: Neither terraform nor tofu found in PATH"
    exit 1
fi

# Check if we're in the right directory
if [ ! -f "main.tf" ]; then
    echo "❌ Error: main.tf not found. Run this script from the test-clouddns-import directory."
    exit 1
fi

# Check for required environment variables
if [ -z "$ANEXIA_TOKEN" ]; then
    echo "❌ Error: ANEXIA_TOKEN environment variable not set."
    echo "   Set it with: export ANEXIA_TOKEN='your-token-here'"
    exit 1
fi

echo "✅ Environment check passed"

# Check if local provider is built
PROVIDER_PATH="$HOME/.terraform.d/plugins/hashicorp.com/anexia-it/anxcloud/0.7.5-dev/linux_amd64/terraform-provider-anxcloud"
if [ ! -f "$PROVIDER_PATH" ]; then
    echo "⚠️  Warning: Local provider not found at $PROVIDER_PATH"
    echo "   Build and install it first:"
    echo "   cd .. && make install"
    echo ""
    read -p "Continue anyway? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
else
    echo "✅ Local provider found"
fi

# Function to show usage
show_usage() {
    echo ""
    echo "Usage: $0 <command>"
    echo ""
    echo "Commands:"
    echo "  build     - Build and install local provider"
    echo "  init      - Initialize Terraform"
    echo "  create    - Create test resources"
    echo "  show      - Show current state and get identifier"
    echo "  test-import - Test import functionality"
    echo "  cleanup   - Destroy all resources"
    echo "  full-test - Run complete test cycle"
    echo ""
}

# Function to get the stable identifier from terraform/tofu show
get_identifier() {
    TF_CLI_CONFIG_FILE=./dev.tfrc $TF_CMD show -json | jq -r '.values.root_module.resources[] | select(.name == "test_record") | .values.identifier'
}

case "${1:-help}" in
    "build")
        echo "🔨 Building and installing local provider..."
        cd ..
        make install
        cd test-clouddns-import
        echo "✅ Provider built and installed"
        ;;

    "init")
        echo "📦 Initializing $TF_CMD..."
        if [ -f "dev.tfrc" ]; then
            echo "ℹ️  Using dev_overrides - skipping init (not needed)"
            echo "✅ $TF_CMD ready with development overrides"
        else
            $TF_CMD init
            echo "✅ $TF_CMD initialized"
        fi
        ;;

    "create")
        echo "🏗️  Creating test resources..."
        TF_CLI_CONFIG_FILE=./dev.tfrc $TF_CMD plan -out=tfplan
        TF_CLI_CONFIG_FILE=./dev.tfrc $TF_CMD apply tfplan
        echo "✅ Resources created"
        echo ""
        echo "🔍 Getting stable identifier..."
        IDENTIFIER=$(TF_CLI_CONFIG_FILE=./dev.tfrc get_identifier)
        echo "📋 Stable Identifier: $IDENTIFIER"
        echo "   Save this for the import test!"
        ;;

    "show")
        echo "📊 Current $TF_CMD state:"
        TF_CLI_CONFIG_FILE=./dev.tfrc $TF_CMD show
        echo ""
        echo "🔍 Stable identifier:"
        IDENTIFIER=$(TF_CLI_CONFIG_FILE=./dev.tfrc get_identifier)
        echo "📋 $IDENTIFIER"
        ;;

    "test-import")
        echo "🧪 Testing import functionality..."

        # Get the identifier
        IDENTIFIER=$(TF_CLI_CONFIG_FILE=./dev.tfrc get_identifier)
        if [ -z "$IDENTIFIER" ] || [ "$IDENTIFIER" = "null" ]; then
            echo "❌ Error: Could not find identifier. Make sure resources are created first."
            exit 1
        fi

        echo "📋 Found identifier: $IDENTIFIER"

        # Remove from state (but keep in API)
        echo "🗑️  Removing from $TF_CMD state..."
        TF_CLI_CONFIG_FILE=./dev.tfrc $TF_CMD state rm anxcloud_dns_record.test_record

        # Import it back
        echo "📥 Importing back using stable identifier..."
        TF_CLI_CONFIG_FILE=./dev.tfrc $TF_CMD import anxcloud_dns_record.test_record "$IMPORT_ID"

        # Verify
        echo "✅ Import completed. Verifying..."
        TF_CLI_CONFIG_FILE=./dev.tfrc $TF_CMD plan

        echo ""
        echo "🎉 Import test completed successfully!"
        echo "   The plan should show no changes if import worked correctly."
        ;;

    "cleanup")
        echo "🧹 Cleaning up test resources..."
        TF_CLI_CONFIG_FILE=./dev.tfrc $TF_CMD destroy -auto-approve
        echo "✅ Cleanup completed"
        ;;

    "show")
        echo "📊 Current $TF_CMD state:"
        $TF_CMD show
        echo ""
        echo "🔍 Stable identifier:"
        IDENTIFIER=$(get_identifier)
        echo "📋 $IDENTIFIER"
        ;;

    "test-import")
        echo "🧪 Testing import functionality..."

        # Get the identifier
        IDENTIFIER=$(get_identifier)
        if [ -z "$IDENTIFIER" ] || [ "$IDENTIFIER" = "null" ]; then
            echo "❌ Error: Could not find identifier. Make sure resources are created first."
            exit 1
        fi

        echo "📋 Found identifier: $IDENTIFIER"

        # Remove from state (but keep in API)
        echo "🗑️  Removing from $TF_CMD state..."
        $TF_CMD state rm anxcloud_dns_record.test_record

        # Import it back using the new format: zone_name/identifier
        ZONE_NAME="test-import-zone.terraform.example"
        IMPORT_ID="${ZONE_NAME}/${IDENTIFIER}"
        echo "📥 Importing back using format: $IMPORT_ID"
        $TF_CMD import anxcloud_dns_record.test_record "$IMPORT_ID"

        # Verify
        echo "✅ Import completed. Verifying..."
        $TF_CMD plan

        echo ""
        echo "🎉 Import test completed successfully!"
        echo "   The plan should show no changes if import worked correctly."
        ;;

    "cleanup")
        echo "🧹 Cleaning up test resources..."
        $TF_CMD destroy -auto-approve
        echo "✅ Cleanup completed"
        ;;

    "full-test")
        echo "🚀 Running full test cycle..."
        echo ""

        # Build
        echo "Step 1: Build provider"
        $0 build
        echo ""

        # Init
        echo "Step 2: Initialize"
        $0 init
        echo ""

        # Create
        echo "Step 3: Create resources"
        $0 create
        echo ""

        # Test import
        echo "Step 4: Test import functionality"
        $0 test-import
        echo ""

        # Cleanup
        echo "Step 5: Cleanup"
        read -p "Press Enter to continue with cleanup..."
        $0 cleanup

        echo ""
        echo "🎉 Full test cycle completed successfully!"
        ;;

    "help"|*)
        show_usage
        ;;
esac
