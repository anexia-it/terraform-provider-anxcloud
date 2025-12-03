#!/bin/bash

# Setup script for CloudDNS import testing
# This script helps configure the development environment

set -e

echo "🔧 Setting up CloudDNS Import Test Environment"
echo "==============================================="

# Check if we're in the right directory
if [ ! -f "main.tf" ] || [ ! -f "dev.tfrc" ]; then
    echo "❌ Error: Please run this script from the test-clouddns-import directory."
    exit 1
fi

# Check for required environment variables
if [ -z "$ANEXIA_TOKEN" ]; then
    echo "⚠️  Warning: ANEXIA_TOKEN environment variable not set."
    echo "   You can set it with: export ANEXIA_TOKEN='your-token-here'"
    echo ""
fi

# Check if jq is available (needed for parsing terraform output)
if ! command -v jq &> /dev/null; then
    echo "❌ Error: jq is required but not installed."
    echo "   Install with: apt-get install jq  (Ubuntu/Debian)"
    echo "   Or: brew install jq  (macOS)"
    exit 1
fi

echo "✅ Prerequisites check passed"

# Check if local provider exists
PROVIDER_DIR="/home/rweselowski/PhpstormProjects/terraform-provider-anxcloud"
if [ ! -d "$PROVIDER_DIR" ]; then
    echo "❌ Error: Provider source directory not found at $PROVIDER_DIR"
    echo "   Please clone the terraform-provider-anxcloud repository there."
    exit 1
fi

echo "✅ Provider source directory found"

# Check if dev.tfrc is properly configured
if ! grep -q "hashicorp.com/anexia-it/anxcloud" dev.tfrc; then
    echo "❌ Error: dev.tfrc is not properly configured."
    echo "   Please check the dev.tfrc file."
    exit 1
fi

echo "✅ Development configuration is valid"

# Clean up any existing terraform state
if [ -d ".terraform" ]; then
    echo "🧹 Cleaning up existing Terraform state..."
    rm -rf .terraform
    rm -f .terraform.lock.hcl
    rm -f tfplan
fi

echo ""
echo "🎉 Setup completed successfully!"
echo ""
echo "Next steps:"
echo "1. Set your ANEXIA_TOKEN: export ANEXIA_TOKEN='your-token-here'"
echo "2. Run the full test: ./test-import.sh full-test"
echo ""
echo "The dev.tfrc file will automatically configure Terraform/OpenTofu to use"
echo "your local provider build without checksum verification issues."