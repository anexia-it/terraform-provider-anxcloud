data "anxcloud_core_location" "anx04" {
  code = "ANX04"
}

resource "anxcloud_vlan" "example" {
  location_id     = data.anxcloud_core_location.anx04.id
  vm_provisioning = true
}
