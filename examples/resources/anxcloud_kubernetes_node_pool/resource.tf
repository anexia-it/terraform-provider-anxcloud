data "anxcloud_kubernetes_cluster" "example" {
  name = "example-cluster"
}

data "anxcloud_vlan" "example" {
  name = "example-network"
}

resource "anxcloud_kubernetes_node_pool" "example" {
  name             = "example-node-pool"
  initial_replicas = 3
  cpus             = 2
  memory_gib       = 4
  operating_system = "Flatcar Linux"
  cluster          = data.anxcloud_kubernetes_cluster.example.id

  networks {
    name            = "internal"
    bandwidth_limit = "1000"
    vlan            = data.anxcloud_vlan.example.id
  }

  disk {
    size_gib = 20
  }
}
