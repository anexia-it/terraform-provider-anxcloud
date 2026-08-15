data "anxcloud_core_location" "e2e" {
  code = var.location_code
}

data "anxcloud_vlan" "internal" {
  id = "097296a1fbb84c42b98cf4c8b3e8b549"
}

data "anxcloud_vlan" "external" {
  id = "b8f265be6fdc4a9cb1d3bc4e8314ceed"
}

data "anxcloud_network_prefix" "internal_v4" {
  id = "9ff417815c944294a88d9fad454d784f"
}

data "anxcloud_network_prefix" "external_v4" {
  id = "2b4ff37dbf8f4dfb80dbd8038fe2318e"
}

resource "anxcloud_kubernetes_cluster" "e2e" {
  name     = "${var.test_id}-cluster"
  location = data.anxcloud_core_location.e2e.id
  version  = var.kubernetes_version

  internal_ipv4_prefix = data.anxcloud_network_prefix.internal_v4.id
  external_ipv4_prefix = data.anxcloud_network_prefix.external_v4.id

  needs_service_vms   = true
  enable_nat_gateways = true
  enable_lbaas        = var.enable_lbaas
  enable_autoscaling  = false

  # These explicitly configured fields exercise the post-create PATCH path.
  cni_plugin                 = "canal"
  external_ip_families       = var.external_ip_families
  enable_oidc_authentication = false

  wait_until_ready = var.wait_until_ready
}

resource "anxcloud_kubernetes_node_pool" "e2e" {
  name             = "${var.test_id}-nodepool"
  cluster          = anxcloud_kubernetes_cluster.e2e.id
  initial_replicas = var.node_replicas
  cpus             = var.node_cpus
  memory_gib       = var.node_memory_gib
  operating_system = "Flatcar Linux"

  sync_source = var.node_sync_source

  # These values are sent with the initial node-pool POST, including zeros.
  cpu_performance_type = "performance"
  autoscaler_enabled   = false
  autoscaler_min_nodes = 0
  autoscaler_max_nodes = 0

  # At least one network must be part of the node-pool creation request.
  networks {
    name            = "internal"
    bandwidth_limit = var.node_pool_network_bandwidth_limit
    vlan            = data.anxcloud_vlan.internal.id
  }

  disk {
    size_gib         = var.node_disk_size_gib
    performance_type = var.node_disk_performance_type
  }
}

# Exercise the standalone resource by adding a second network after the node
# pool has been created with its required inline network.
resource "anxcloud_kubernetes_node_pool_network" "external" {
  name            = "external"
  node_pool       = anxcloud_kubernetes_node_pool.e2e.id
  vlan            = data.anxcloud_vlan.external.id
  bandwidth_limit = var.node_pool_network_bandwidth_limit
}
