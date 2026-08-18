data "anxcloud_network_prefix" "by_id" {
  id = "0123456789abcdef0123456789abcdef"
}

data "anxcloud_network_prefix" "by_cidr" {
  cidr = "10.0.0.0/24"
}
