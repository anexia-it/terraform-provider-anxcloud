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


output "cluster-admin-config" {
  value     = anxcloud_kubernetes_kubeconfig.cluster-admin.raw
  sensitive = true
}

output "cluster-admin-token" {
  value     = anxcloud_kubernetes_kubeconfig.cluster-admin.token
  sensitive = true
}

output "load_balancer_smoke_test" {
  description = "Traefik load-balancer addresses. Open the IP or hostname over HTTP and expect 'Anexia Kubernetes LBaaS works'."
  value = {
    namespace = kubernetes_namespace_v1.anexia.metadata[0].name
    ingress   = kubernetes_ingress_v1.smoke_app.metadata[0].name
    addresses = try(data.kubernetes_service_v1.traefik.status[0].load_balancer[0].ingress, [])
  }
}
