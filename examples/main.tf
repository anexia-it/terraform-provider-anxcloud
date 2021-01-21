terraform {
  required_providers {
    anxcloud = {
      versions = ["0.2.3"]
      source   = "hashicorp.com/anexia-it/anxcloud"
    }
  }
}

provider "anxcloud" {}

data "anxcloud_core_locations" "example" {}