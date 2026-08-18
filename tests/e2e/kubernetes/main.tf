resource "anxcloud_kubernetes_cluster" "e2e" {
  name            = "${var.test_id}-cluster"
  location        = data.anxcloud_core_location.e2e.id
  version         = var.kubernetes_version
  api_environment = var.kubernetes_api_environment

  internal_ipv4_prefix = data.anxcloud_network_prefix.internal_v4.id
  external_ipv4_prefix = data.anxcloud_network_prefix.external_v4.id

  needs_service_vms   = true
  enable_nat_gateways = true
  enable_lbaas        = var.enable_lbaas
  enable_autoscaling  = true

  cni_plugin           = "canal"
  external_ip_families = var.external_ip_families

  apiserver_allowlist = ["1.2.3.4"]

  enable_oidc_authentication = true
  oidc_client_id             = "1234"
  oidc_issuer_url            = "https://localhost:8080"
  oidc_groups_claim          = "groups"
  oidc_username_claim        = "user"
  oidc_extra_scopes          = "user,group"
  oidc_groups_prefix         = "e2e_"
  oidc_required_claim        = "group"
  oidc_username_prefix       = "e2e_user"

  maintenance_window_start_time = "Mon 00:00"
  maintenance_window_duration   = "2h"

  wait_until_ready = var.wait_until_ready
}

resource "anxcloud_kubernetes_node_pool" "e2e" {
  name             = "${var.test_id}-nodepool"
  cluster          = anxcloud_kubernetes_cluster.e2e.id
  initial_replicas = var.node_replicas
  cpus             = var.node_cpus
  memory_gib       = var.node_memory_gib
  operating_system = "Flatcar Linux"
  api_environment  = var.kubernetes_api_environment

  sync_source = var.node_sync_source

  # These values are sent with the initial node-pool POST, including zeros.
  cpu_performance_type = "performance"
  autoscaler_enabled   = true
  autoscaler_min_nodes = 1
  autoscaler_max_nodes = 4

  # All node-pool networks are owned by the node pool and sent in its POST.
  networks {
    name            = "internal"
    bandwidth_limit = var.node_pool_network_bandwidth_limit
    vlan            = data.anxcloud_vlan.internal.id
  }

  # Just an example on how to add multiple networks
  # networks {
  #   name            = "external"
  #   bandwidth_limit = var.node_pool_network_bandwidth_limit
  #   vlan            = data.anxcloud_vlan.external.id
  # }

  disk {
    size_gib         = var.node_disk_size_gib
    performance_type = var.node_disk_performance_type
  }
}

resource "anxcloud_kubernetes_kubeconfig" "cluster-admin" {
  cluster         = anxcloud_kubernetes_cluster.e2e.id
  api_environment = var.kubernetes_api_environment
}
