# create a specific address within a VLAN
resource "anxcloud_ip_address" "example_specific" {
  network_prefix_id = anxcloud_network_prefix.example.id
  address           = "10.20.30.1"
}

# reserve an address in a specific prefix
resource "anxcloud_ip_address" "example_prefix_v4" {
  vlan_id           = anxcloud_vlan.example.id
  network_prefix_id = anxcloud_network_prefix.example.id
}

# reserve an address in any v4 prefix available in the specified VLAN
resource "anxcloud_ip_address" "example_version_v4" {
  vlan_id = anxcloud_vlan.example.id
  version = 4
}

# reserve an address in any v6 prefix available in the specified VLAN
resource "anxcloud_ip_address" "example_version_v6" {
  vlan_id = anxcloud_vlan.example.id
  version = 6
}

# reserve an address in any prefix available in the specified VLAN
resource "anxcloud_ip_address" "example_any_in_vlan" {
  vlan_id = anxcloud_vlan.example.id
}
