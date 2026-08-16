data "anxcloud_core_location" "anx04" {
  code = "ANX04"
}

data "anxcloud_vlan" "internal" {
  id = "example"
}

data "anxcloud_vlan" "external" {
  id = "example"
}

data "anxcloud_kubernetes_cluster" "example" {
  id = "example"
}

resource "anxcloud_kubernetes_node_pool" "e2e" {
  name    = "example-nodepool"
  cluster = data.anxcloud_kubernetes_cluster.example.id

  # Nodepools are represented as MachineDeployments (Cluster API) in kubernetes.
  # We offer to manage them either through the Anexia engine or the Kubernetes cluster
  sync_source = "engine"
  # sync_source = "Cluster"

  # Sizing parameters
  operating_system     = "Flatcar Linux"
  initial_replicas     = 1
  cpus                 = 4
  cpu_performance_type = "performance"
  memory_gib           = 8
  disk {
    size_gib         = 30
    performance_type = "ENT2"
  }

  # Finer autoscaling configurations happen on the cluster-autoscaler level.
  # See https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/FAQ.md
  autoscaler_enabled   = true
  autoscaler_min_nodes = 1
  autoscaler_max_nodes = 4

  # All node-pool networks are owned by the node pool and sent in its POST.
  networks {
    name            = "internal"
    bandwidth_limit = 1000
    vlan            = data.anxcloud_vlan.internal.id
  }

  # Just an example on how to add multiple networks
  # networks {
  #   name            = "external"
  #   bandwidth_limit = 1000
  #   vlan            = data.anxcloud_vlan.external.id
  # }
}
