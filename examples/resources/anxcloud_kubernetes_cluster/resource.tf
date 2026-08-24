data "anxcloud_core_location" "anx04" {
  code = "ANX04"
}

data "anxcloud_vlan" "internal" {
  id = "example"
}

data "anxcloud_vlan" "external" {
  id = "example"
}

# create a kubernetes cluster with servicevm loadbalancer and IPv4 only networks
resource "anxcloud_kubernetes_cluster" "example" {
  name     = "example-cluster"
  location = data.anxcloud_core_location.anx04.id
  version  = "1.35"

  internal_ipv4_prefix = data.anxcloud_network_prefix.internal_v4.id
  external_ipv4_prefix = data.anxcloud_network_prefix.external_v4.id

  needs_service_vms   = true
  enable_nat_gateways = true
  enable_lbaas        = true
  enable_autoscaling  = true

  cni_plugin                 = "canal"
  external_ip_families       = "IPv4"
  enable_oidc_authentication = false

  # Recommended: wait for cluster reconciliation to complete before Terraform returns.
  wait_until_ready = true
}

resource "anxcloud_kubernetes_kubeconfig" "cluster-admin" {
  cluster = anxcloud_kubernetes_cluster.example.id
}
