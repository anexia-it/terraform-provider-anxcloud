data "anxcloud_core_location" "anx04" {
  code = "ANX04"
}

data "anxcloud_template" "anx04" {
  location_id   = data.anxcloud_core_location.anx04.id
  template_type = "templates"
}

locals {
  debian_11_templates = values({
    for i, template in data.anxcloud_template.anx04.templates : tostring(i) => template
    if substr(template.name, 0, 9) == "Debian 11"
  })
}

resource "anxcloud_vlan" "example" {
  count           = 2
  location_id     = data.anxcloud_core_location.anx04.id
  vm_provisioning = true
}

resource "anxcloud_virtual_server" "example" {
  hostname      = "example-terraform"
  location_id   = data.anxcloud_core_location.anx04.id
  template_id   = local.debian_11_templates[0].id
  template_type = "templates"

  cpus   = 4
  memory = 4096

  ssh_key = file("~/.ssh/id_rsa.pub")

  # set two network interfaces
  # NIC 1
  network {
    vlan_id  = anxcloud_vlan.example[0].id
    nic_type = "vmxnet3"
  }

  # NIC 2
  network {
    vlan_id  = anxcloud_vlan.example[1].id
    nic_type = "vmxnet3"
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
