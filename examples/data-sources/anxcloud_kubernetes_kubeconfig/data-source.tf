data "anxcloud_kubernetes_cluster" "example" {
  name            = "example-cluster"
  api_environment = "prod"
}

data "anxcloud_kubernetes_kubeconfig" "example" {
  cluster         = data.anxcloud_kubernetes_cluster.example.id
  api_environment = "prod"
}

provider "kubernetes" {
  host                   = data.anxcloud_kubernetes_kubeconfig.example.host
  token                  = data.anxcloud_kubernetes_kubeconfig.example.token
  cluster_ca_certificate = data.anxcloud_kubernetes_kubeconfig.example.cluster_ca_certificate
}
