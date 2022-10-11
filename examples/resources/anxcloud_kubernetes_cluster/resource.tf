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
