data "anxcloud_ip_address" "example" {
  address = "10.244.2.50"
}

// This data-source can be used to create virtual servers.
// The following example is incomplete and won't work as is!
// Check out the anxcloud_virtual_server documentation for a full example.
resource "anxcloud_virtual_server" "example" {

  network {
    vlan_id  = data.anxcloud_ip_address.example.vlan_id
    ips      = [data.anxcloud_ip_address.example.address]
    nic_type = "vmxnet3"
  }

}
