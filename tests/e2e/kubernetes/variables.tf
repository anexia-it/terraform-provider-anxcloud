variable "test_id" {
  type        = string
  description = "Unique lowercase identifier used in names and descriptions for this E2E run."

  validation {
    condition = (
      length(var.test_id) >= 2 &&
      length(var.test_id) <= 25 &&
      can(regex("^[a-z0-9][a-z0-9-]*[a-z0-9]$", var.test_id))
    )
    error_message = "test_id must be 2-25 lowercase letters, numbers, or hyphens, and may not start or end with a hyphen."
  }
}

variable "location_code" {
  type        = string
  description = "Anexia location code in which the test resources are created."
  default     = "ANX25"
}

variable "wait_until_ready" {
  type        = bool
  description = "Wait for the cluster to reach reconciliation state 0 before creating the node pool."
  default     = true
}

variable "kubernetes_version" {
  type        = string
  description = "Kubernetes minor version sent during cluster creation."
  default     = "1.35"
}

variable "external_ip_families" {
  type        = string
  description = "External cluster IP families."
  default     = "IPv4"

  validation {
    condition     = contains(["IPv4", "Dualstack"], var.external_ip_families)
    error_message = "external_ip_families must be IPv4 or Dualstack."
  }
}

variable "enable_lbaas" {
  type        = bool
  description = "Whether LBaaS is enabled on the cluster. Change this to test an in-place cluster PATCH."
  default     = true
}

variable "node_replicas" {
  type        = number
  description = "Desired number of nodes. Change this to test an in-place node-pool PATCH."
  default     = 2
}

variable "node_sync_source" {
  type        = string
  description = "Source of truth for node-pool configuration. Defaults to Engine; set to Cluster to hand control to the cluster."
  default     = "engine"

  validation {
    condition     = contains(["engine", "cluster", "Cluster"], var.node_sync_source)
    error_message = "node_sync_source must be engine or Cluster."
  }
}

variable "node_cpus" {
  type        = number
  description = "CPU count per node."
  default     = 2
}

variable "node_memory_gib" {
  type        = number
  description = "Memory per node in GiB."
  default     = 4
}

variable "node_disk_size_gib" {
  type        = number
  description = "Primary disk size per node in GiB."
  default     = 20
}

variable "node_disk_performance_type" {
  type        = string
  description = "Primary node disk performance type."
  default     = "STD1"
}

variable "node_pool_network_bandwidth_limit" {
  type        = string
  description = "Bandwidth limit identifier used by both node-pool networks."
  default     = "1000"
}

variable "anexia_token" {
  type        = string
  description = "Anexia API token written to the anexia/anexia-credentials Kubernetes secret. Set it through TF_VAR_anexia_token."
  sensitive   = true
  nullable    = false

  validation {
    condition     = length(var.anexia_token) > 0
    error_message = "anexia_token must not be empty."
  }
}

variable "generic_helm_chart_version" {
  type        = string
  description = "Version of the Anexia ks-generic-helmchart used for the ingress smoke test."
  default     = "0.1.31"
}
