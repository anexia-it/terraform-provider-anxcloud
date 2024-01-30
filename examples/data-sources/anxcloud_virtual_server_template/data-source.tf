data "anxcloud_virtual_server_template" "debian11" {
  name     = "Debian 11"
  location = data.anxcloud_core_location.anx04.id
}
