terraform {
  required_providers {
    anxcloud = {
      source  = "anexia-it/anxcloud"
      version = "~> 0.3"
    }
  }
}

// Authentication via environment variable is strongly advised:
// export ANEXIA_TOKEN='<token>'

// Alternatively, but NOT recommended:
provider "anxcloud" {
  token = "<token>"
}


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
  description_customer = "example-terraform"
  location_id          = data.anxcloud_core_location.anx04.id
  vm_provisioning      = true
}

resource "anxcloud_network_prefix" "example" {
  vlan_id     = anxcloud_vlan.example.id
  location_id = data.anxcloud_core_location.anx04.id
  ip_version  = 4
  type        = 0
  netmask     = 29
}

resource "anxcloud_virtual_server" "example" {
  hostname      = "example-terraform"
  location_id   = data.anxcloud_core_location.anx04.id
  template_id   = local.debian_11_templates[0].id
  template_type = "templates"

  cpus   = 4
  memory = 4096

  ssh_key = file("~/.ssh/id_rsa.pub")

  network {
    vlan_id  = anxcloud_vlan.example.id
    nic_type = "vmxnet3"
  }

  # Disk 1
  disk {
    disk_gb   = 100
    disk_type = "STD1"
  }

  dns = ["8.8.8.8"]

  depends_on = [
    anxcloud_network_prefix.example
  ]
}