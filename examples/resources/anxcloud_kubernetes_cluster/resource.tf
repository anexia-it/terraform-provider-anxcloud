terraform {
  required_providers {
    anxcloud = {
      source = "hashicorp.com/anexia-it/anxcloud"
      version = "0.11.1"
    }
  }
}

provider "anxcloud" {}

#####################################################
##  Create cluster within existing infrastructure  ##
#####################################################

data "anxcloud_core_location" "anx25" {
  code = "ANX25"
}

###################### VLANS ########################


data "anxcloud_vlan" "internal" {
  # name = "VLAN1234"
  id = "573674b9007845f498e787fec41e04cc"
}

data "anxcloud_vlan" "external" {
  # name = "VLAN1234"
  id = "1fb1874ec9c3404f9d69e76dc64f7505"
}

# resource "anxcloud_vlan" "internal" {
#   location_id          = data.anxcloud_core_location.anx25.id
#   description_customer = "internal"
#   vm_provisioning      = true
# }
#
# resource "anxcloud_vlan" "external" {
#   location_id          = data.anxcloud_core_location.anx25.id
#   description_customer = "external"
#   vm_provisioning      = true
# }


#################### PREFIXES #######################

data "anxcloud_network_prefix" "internal_v4" {
  id = "fb539f737aac46d5894e069d06fb3aca"
}

data "anxcloud_network_prefix" "external_v4" {
  id = "abc51f3d0bf94e089c759060e9db6b16"
}


################## CLUSTER #####################

resource "anxcloud_kubernetes_cluster" "foo" {
  name     = "foobar2"
  location = data.anxcloud_core_location.anx25.id
  version = "1.35"
  external_ip_families = "IPv4"

  internal_ipv4_prefix = data.anxcloud_network_prefix.internal_v4.id
  external_ipv4_prefix = data.anxcloud_network_prefix.external_v4.id
  wait_until_ready = true
}
