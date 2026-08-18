data "anxcloud_core_location" "e2e" {
  code = var.location_code
}

data "anxcloud_vlan" "internal" {
  id = var.internal_ipv4_vlan
}

data "anxcloud_vlan" "external" {
  id = var.external_ipv4_vlan
}

data "anxcloud_network_prefix" "internal_v4" {
  id = var.internal_ipv4_prefix
}

data "anxcloud_network_prefix" "external_v4" {
  id = var.external_ipv4_prefix
}