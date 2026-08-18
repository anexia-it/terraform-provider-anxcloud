terraform {
  required_version = ">= 1.5.0"

  required_providers {
    anxcloud = {
      source  = "hashicorp.com/anexia-it/anxcloud"
      version = "0.11.1"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "3.2.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "3.2.1"
    }
  }
}

provider "anxcloud" {}
