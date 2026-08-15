terraform {
  required_version = ">= 1.5.0"

  required_providers {
    anxcloud = {
      source  = "hashicorp.com/anexia-it/anxcloud"
      version = "0.3.1"
    }
  }
}

provider "anxcloud" {}
