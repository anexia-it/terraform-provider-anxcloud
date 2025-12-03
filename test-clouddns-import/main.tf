terraform {
  required_providers {
    anxcloud = {
      source  = "hashicorp.com/anexia-it/anxcloud"
      version = "0.3.1"
    }
  }
}

# Use local provider build for testing our implementation
# Build the provider first: cd .. && make install
provider "anxcloud" {
  # token = "your-api-token-here"  # or set ANEXIA_TOKEN env var
}

# Create a DNS zone for testing
resource "anxcloud_dns_zone" "test_zone" {
  name         = "test-import-zone.terraform.example"
  is_master    = true
  dns_sec_mode = "unvalidated"
  admin_email  = "admin@example.com"
  refresh      = 3600
  retry        = 1800
  expire       = 604800
  ttl          = 86400
}

# Create a DNS record that we can test importing
resource "anxcloud_dns_record" "test_record" {
  name      = "test-record"
  zone_name = anxcloud_dns_zone.test_zone.name
  type      = "A"
  rdata     = "192.168.1.100"
  ttl       = 300
}

# Output the stable identifier for easy access
output "dns_record_identifier" {
  value       = anxcloud_dns_record.test_record.identifier
  description = "Stable identifier for the DNS record - use this for import testing"
}

output "dns_record_id" {
  value       = anxcloud_dns_record.test_record.id
  description = "Terraform resource ID"
}