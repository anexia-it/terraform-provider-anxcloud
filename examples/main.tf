terraform {
  required_providers {
    anxcloud = {
      versions = ["0.1.0"]
      source   = "hashicorp.com/anexia-it/anxcloud"
    }
  }
}

provider "anxcloud" {}

resource "anxcloud_virtual_server" "example" {
  location_id   = "52b5f6b2fd3a4a7eaaedf1a7c019e9ea"
  template_id   = "12c28aa7-604d-47e9-83fb-5f1d1f1837b3"
  template_type = "templates"
  hostname      = "example-terraform"
  cpus          = 4
  disk          = 50
  memory        = 4096
  password      = "flatcar#1234$%"

  network {
    vlan_id  = "ff70791b398e4ab29786dd34f211694c"
    nic_type = "vmxnet3"
  }

  network {
    vlan_id  = "ff70791b398e4ab29786dd34f211694c"
    nic_type = "vmxnet3"
  }

  dns = ["8.8.8.8"]

}

resource "anxcloud_vlan" "example" {
  location_id = "52b5f6b2fd3a4a7eaaedf1a7c019e9ea"
  vm_provisioning = true
  description_customer = "sample vlan"
}
