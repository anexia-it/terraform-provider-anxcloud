data "anxcloud_core_location" "anx04" {
  code = "ANX04"
}

resource "anxcloud_vlan" "example" {
  location_id = data.anxcloud_core_location.anx04.id
}

resource "anxcloud_network_prefix" "example" {
  vlan_id     = anxcloud_vlan.example.id
  location_id = data.anxcloud_core_location.anx04.id
  ip_version  = 4
  type        = 0
  netmask     = 29
}
