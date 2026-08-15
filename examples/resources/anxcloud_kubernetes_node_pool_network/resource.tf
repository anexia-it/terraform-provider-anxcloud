data "anxcloud_vlan" "node_pool_network" {
  id = "existing-vlan-identifier"
}

resource "anxcloud_kubernetes_node_pool_network" "example" {
  name            = "storage"
  node_pool       = anxcloud_kubernetes_node_pool.example.id
  vlan            = data.anxcloud_vlan.node_pool_network.id
  bandwidth_limit = "1000"
}
