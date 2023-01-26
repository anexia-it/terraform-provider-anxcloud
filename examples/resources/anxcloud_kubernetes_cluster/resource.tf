data "anxcloud_core_location" "anx04" {
  code = "ANX04"
}

resource "anxcloud_kubernetes_cluster" "example" {
  name     = "example-cluster"
  location = data.anxcloud_core_location.anx04.id
}

resource "anxcloud_kubernetes_node_pool" "example" {
  name             = "example-node-pool"
  initial_replicas = 3
  cpus             = 2
  memory_gib       = 4
  operating_system = "Flatcar Linux"
  cluster          = anxcloud_kubernetes_cluster.example.id

  disk {
    size_gib = 20
  }
}



#####################################################
##  Create cluster within existing infrastructure  ##
#####################################################

data "anxcloud_core_location" "anx04" {
  code = "ANX04"
}

###################### VLANS ########################

resource "anxcloud_vlan" "internal" {
  location_id          = data.anxcloud_core_location.anx04.id
  description_customer = "internal"
  vm_provisioning      = true
}

resource "anxcloud_vlan" "external" {
  location_id          = data.anxcloud_core_location.anx04.id
  description_customer = "external"
  vm_provisioning      = true
}


#################### PREFIXES #######################

resource "anxcloud_network_prefix" "internal_v4" {
  description_customer = "internal v4"
  type                 = 1
  netmask              = 24
  ip_version           = 4
  create_empty         = true
  location_id          = data.anxcloud_core_location.anx04.id
  vlan_id              = anxcloud_vlan.internal.id
}

resource "anxcloud_network_prefix" "external_v4" {
  description_customer = "external v4"
  type                 = 0
  netmask              = 28
  ip_version           = 4
  create_empty         = true
  location_id          = data.anxcloud_core_location.anx04.id
  vlan_id              = anxcloud_vlan.external.id
}

resource "anxcloud_network_prefix" "external_v6" {
  description_customer = "external v6"
  type                 = 0
  netmask              = 64
  ip_version           = 6
  create_empty         = true
  location_id          = data.anxcloud_core_location.anx04.id
  vlan_id              = anxcloud_vlan.external.id
}

################## CLUSTER #####################

resource "anxcloud_kubernetes_cluster" "foo" {
  name     = "foo"
  location = data.anxcloud_core_location.anx04.id

  internal_ipv4_prefix = anxcloud_network_prefix.internal_v4.id
  external_ipv4_prefix = anxcloud_network_prefix.external_v4.id
  external_ipv6_prefix = anxcloud_network_prefix.external_v6.id
}
