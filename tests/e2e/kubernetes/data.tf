data "anxcloud_core_location" "e2e" {
  code = var.location_code
}

data "anxcloud_vlan" "internal" {
  id = "f19894e3d89c43c5a7805ff1d6f405b8"
}

data "anxcloud_vlan" "external" {
  id = "271f0f1ef82c4bb6b9d1d6906d03e06b"
}

data "anxcloud_network_prefix" "internal_v4" {
  id = "028f2ea718354cd0a4c48c216f5c5b5c"
}

data "anxcloud_network_prefix" "external_v4" {
  id = "59c75b9c1cd24151ae9cb956834119dd"
}