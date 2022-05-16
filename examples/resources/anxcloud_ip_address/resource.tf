data "anxcloud_core_location" "anx04" {
  code = "ANX04"
}

resource "anxcloud_ip_address" "example" {
  network_prefix_id = var.network_prefix_id
  address           = "10.20.30.1"
}
