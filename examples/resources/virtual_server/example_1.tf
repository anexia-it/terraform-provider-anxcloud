data "anxcloud_core_location" "anx04" {
  code = "ANX04"
}

resource "anxcloud_vlan" "example" {
  location_id          = data.anxcloud_core_location.anx04.id
  vm_provisioning      = true
  description_customer = "example-terraform"
}

resource "anxcloud_network_prefix" "v4" {
  vlan_id              = anxcloud_vlan.example.id
  location_id          = data.anxcloud_core_location.anx04.id
  ip_version           = 4
  netmask              = 30
  description_customer = "example-terraform"
}

resource "anxcloud_network_prefix" "v6" {
  vlan_id              = anxcloud_vlan.example.id
  location_id          = data.anxcloud_core_location.anx04.id
  ip_version           = 6
  netmask              = 126
  description_customer = "example-terraform"
}

resource "anxcloud_ip_address" "v4" {
  address           = cidrhost(anxcloud_network_prefix.v4.cidr, 2)
  network_prefix_id = anxcloud_network_prefix.v4.id
}

resource "anxcloud_ip_address" "v6" {
  address           = cidrhost(anxcloud_network_prefix.v6.cidr, 2)
  network_prefix_id = anxcloud_network_prefix.v6.id
}

resource "anxcloud_virtual_server" "example" {
  hostname    = "example-terraform"
  location_id = data.anxcloud_core_location.anx04.id
  template    = "Debian 11"

  cpus   = 4
  memory = 4096

  ssh_key = file("~/.ssh/id_rsa.pub")

  # define bootstrap script
  # e.g. install software
  script = <<-EOT
    #!/bin/bash

    # install nginx server
    apt update && apt install -y nginx
    EOT

  # Set network interface
  network {
    vlan_id  = anxcloud_vlan.example.id
    ips      = [anxcloud_ip_address.v4.id, anxcloud_ip_address.v6.id]
    nic_type = "virtio"
  }

  # Disk 1
  disk {
    disk_gb   = 100
    disk_type = "STD1"
  }

  # Disk 2
  disk {
    disk_gb   = 200
    disk_type = "STD1"
  }

  dns = ["8.8.8.8"]
}
