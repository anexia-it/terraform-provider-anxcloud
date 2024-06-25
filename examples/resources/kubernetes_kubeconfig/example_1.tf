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

resource "anxcloud_kubernetes_kubeconfig" "example" {
  cluster = anxcloud_kubernetes_cluster.example.id
}

provider "kubernetes" {
  host                   = anxcloud_kubernetes_kubeconfig.example.host
  token                  = anxcloud_kubernetes_kubeconfig.example.token
  cluster_ca_certificate = anxcloud_kubernetes_kubeconfig.example.cluster_ca_certificate
}

resource "kubernetes_namespace" "example" {
  metadata {
    name = "example"
  }
}
