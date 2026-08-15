provider "kubernetes" {
  host                   = anxcloud_kubernetes_kubeconfig.cluster-admin.host
  token                  = anxcloud_kubernetes_kubeconfig.cluster-admin.token
  cluster_ca_certificate = anxcloud_kubernetes_kubeconfig.cluster-admin.cluster_ca_certificate
}

provider "helm" {
  kubernetes = {
    host                   = anxcloud_kubernetes_kubeconfig.cluster-admin.host
    token                  = anxcloud_kubernetes_kubeconfig.cluster-admin.token
    cluster_ca_certificate = anxcloud_kubernetes_kubeconfig.cluster-admin.cluster_ca_certificate
  }
}

locals {
  smoke_app_labels = {
    "app.kubernetes.io/name"     = "anexia-lbaas-smoke-test"
    "app.kubernetes.io/instance" = var.test_id
  }
}

resource "kubernetes_namespace_v1" "anexia" {
  metadata {
    name = "anexia"
  }

  depends_on = [anxcloud_kubernetes_node_pool.e2e]
}

resource "kubernetes_secret_v1" "anexia_credentials" {
  metadata {
    name      = "anexia-credentials"
    namespace = kubernetes_namespace_v1.anexia.metadata[0].name
  }

  data = {
    token = var.anexia_token
  }

  type = "Opaque"
}

resource "helm_release" "traefik" {
  name       = "anexia-ingress"
  repository = "oci://anx-cr.io/se/charts"
  chart      = "ks-generic-helmchart"
  version    = var.generic_helm_chart_version
  namespace  = kubernetes_namespace_v1.anexia.metadata[0].name

  atomic          = true
  cleanup_on_fail = true
  timeout         = 900
  wait            = true

  values = [yamlencode({
    anxCsiDriver = {
      enabled = false
    }
    "external-dns" = {
      enabled = false
    }
    "cert-manager" = {
      enabled = false
    }
    "cert-manager-webhook-anexia" = {
      enabled = false
    }
    clusterIssuer = {
      enabled = false
    }
    traefik = {
      enabled          = true
      fullnameOverride = "traefik-ingress"
      deployment = {
        enabled = true
      }
      ingressRoute = {
        dashboard = {
          enabled = false
        }
      }
      providers = {
        kubernetesCRD = {
          enabled = true
        }
        kubernetesIngress = {
          enabled = true
        }
      }
      service = {
        type = "LoadBalancer"
        spec = {
          externalTrafficPolicy = "Local"
        }
      }
    }
    headlamp = {
      enabled = false
    }
    fluentd = {
      enabled = false
    }
    reloader = {
      enabled = false
    }
    velero = {
      enabled = false
    }
    "prometheus-stack" = {
      enabled = false
    }
    "external-secrets" = {
      enabled = false
    }
  })]
}

resource "kubernetes_config_map_v1" "smoke_app" {
  metadata {
    name      = "lbaas-smoke-test"
    namespace = kubernetes_namespace_v1.anexia.metadata[0].name
  }

  data = {
    "index.html" = <<-HTML
      <!doctype html>
      <html>
        <head><title>Anexia Kubernetes LBaaS smoke test</title></head>
        <body><h1>Anexia Kubernetes LBaaS works</h1></body>
      </html>
    HTML
  }
}

resource "kubernetes_deployment_v1" "smoke_app" {
  metadata {
    name      = "lbaas-smoke-test"
    namespace = kubernetes_namespace_v1.anexia.metadata[0].name
    labels    = local.smoke_app_labels
  }

  spec {
    replicas = 1

    selector {
      match_labels = local.smoke_app_labels
    }

    template {
      metadata {
        labels = local.smoke_app_labels
      }

      spec {
        container {
          name  = "nginx"
          image = "nginx:1.29-alpine"

          port {
            name           = "http"
            container_port = 80
          }

          readiness_probe {
            http_get {
              path = "/"
              port = "http"
            }
          }

          volume_mount {
            name       = "content"
            mount_path = "/usr/share/nginx/html"
            read_only  = true
          }
        }

        volume {
          name = "content"

          config_map {
            name = kubernetes_config_map_v1.smoke_app.metadata[0].name
          }
        }
      }
    }
  }
}

resource "kubernetes_service_v1" "smoke_app" {
  metadata {
    name      = "lbaas-smoke-test"
    namespace = kubernetes_namespace_v1.anexia.metadata[0].name
  }

  spec {
    selector = local.smoke_app_labels

    port {
      name        = "http"
      port        = 80
      target_port = "http"
    }
  }
}

resource "kubernetes_ingress_v1" "smoke_app" {
  metadata {
    name      = "lbaas-smoke-test"
    namespace = kubernetes_namespace_v1.anexia.metadata[0].name
  }

  spec {
    ingress_class_name = "traefik-ingress"

    rule {
      http {
        path {
          path      = "/"
          path_type = "Prefix"

          backend {
            service {
              name = kubernetes_service_v1.smoke_app.metadata[0].name
              port {
                number = 80
              }
            }
          }
        }
      }
    }
  }

  depends_on = [helm_release.traefik]
}

data "kubernetes_service_v1" "traefik" {
  metadata {
    name      = "traefik-ingress"
    namespace = kubernetes_namespace_v1.anexia.metadata[0].name
  }

  depends_on = [helm_release.traefik]
}
