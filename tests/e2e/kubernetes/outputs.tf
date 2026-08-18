output "cluster" {
  description = "Created cluster identifier and last observed reconciliation state."
  value = {
    id         = anxcloud_kubernetes_cluster.e2e.id
    name       = anxcloud_kubernetes_cluster.e2e.name
    state      = anxcloud_kubernetes_cluster.e2e.state
    state_text = anxcloud_kubernetes_cluster.e2e.state_text
  }
}

output "node_pool" {
  description = "Created node-pool identifier and last observed reconciliation state."
  value = {
    id         = anxcloud_kubernetes_node_pool.e2e.id
    name       = anxcloud_kubernetes_node_pool.e2e.name
    state      = anxcloud_kubernetes_node_pool.e2e.state
    state_text = anxcloud_kubernetes_node_pool.e2e.state_text
  }
}

output "network" {
  description = "Existing VLAN and prefix inputs attached to the cluster and node pool."
  value = {
    internal_vlan        = data.anxcloud_vlan.internal.id
    external_vlan        = data.anxcloud_vlan.external.id
    internal_ipv4_prefix = data.anxcloud_network_prefix.internal_v4.id
    external_ipv4_prefix = data.anxcloud_network_prefix.external_v4.id
  }
}
