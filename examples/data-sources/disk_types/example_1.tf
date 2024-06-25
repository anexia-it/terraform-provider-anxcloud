data "anxcloud_core_location" "anx04" {
  code = "ANX04"
}

data "anxcloud_disk_types" "example" {
  location_id = data.anxcloud_core_location.anx04.id
}
