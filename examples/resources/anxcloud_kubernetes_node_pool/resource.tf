data "anxcloud_kubernetes_cluster" "example" {
  name = "example-cluster"
}

resource "anxcloud_kubernetes_node_pool" "example" {
  name             = "example-node-pool"
  initial_replicas = 3
  cpus             = 2
  memory_gib       = 4
  operating_system = "Flatcar Linux"
  cluster          = data.anxcloud_kubernetes_cluster.example.id

  disk {
    size_gib = 20
  }
}
