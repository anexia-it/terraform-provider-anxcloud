data "anxcloud_kubernetes_cluster" "example" {
  name = "example-cluster"
}

resource "anxcloud_kubernetes_kubeconfig" "example" {
  cluster = data.anxcloud_kubernetes_cluster.example.id
}

resource "local_file" "kubeconfig" {
  content  = anxcloud_kubernetes_kubeconfig.example.raw
  filename = "${path.module}/kubeconfig"
}
