terraform {
  required_providers {
    anxcloud = {
      versions = ["0.3.1"]
      source   = "hashicorp.com/anexia-it/anxcloud"
    }
  }
}

provider "anxcloud" {}

locals {
  disk_types = {
  for obj in data.anxcloud_disk_types.example.types : obj.id => obj
  }

  template_id = "12c28aa7-604d-47e9-83fb-5f1d1f1837b3"
  location_id = "52b5f6b2fd3a4a7eaaedf1a7c019e9ea"
}

resource "anxcloud_tag" "example" {
  name = "example"
  service_id = "ff543fc08b3149ee9a8c50ee018b15a6"
}

resource "anxcloud_tag" "example2" {
  name = "example2"
  service_id = "ff543fc08b3149ee9a8c50ee018b15a6"
}

data "anxcloud_template" "example" {
  location_id = local.location_id
  template_type = "templates"
}

data "anxcloud_disk_types" "example" {
  location_id = local.location_id
}

resource "anxcloud_vlan" "example" {
  location_id = local.location_id
  vm_provisioning = true
  description_customer = "terraform vlan test"
}

resource "anxcloud_network_prefix" "example" {
  vlan_id = anxcloud_vlan.example.id
  location_id = local.location_id
  ip_version = 4
  type = 0
  netmask = 29
  vm_provisioning = true
  description_customer = "terraform prefix test"
}

resource "anxcloud_virtual_server" "example" {
  location_id   = local.location_id
  template_id   = "12c28aa7-604d-47e9-83fb-5f1d1f1837b3"
  template_type = "templates"
  hostname      = "example-terraform9"
  cpus          = 2
  memory        = 2048
  password      = "flatcar#1234$%"
  cpu_performance_type = "standard"

  network {
    vlan_id  = anxcloud_vlan.example.id
    nic_type = "vmxnet3"
  }

  disk {
    disk_gb       = 70
    disk_type     = local.disk_types.STD6.id
  }

  dns = ["8.8.8.8"]
  force_restart_if_needed = true
  critical_operation_confirmed = true

  tags = [
    anxcloud_tag.example.name,
    anxcloud_tag.example2.name
  ]

  depends_on = [
    anxcloud_network_prefix.example,
  ]
}

data "anxcloud_core_locations" "example" {}

data "anxcloud_tags" "example" {}

data "anxcloud_cpu_performance_types" "example" {}

data "anxcloud_vsphere_locations" "example" {}
