package anxcloud

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func schemaKubernetesCluster() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"api_environment": schemaKubernetesAPIEnvironment(),
		"id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Cluster identifier.",
		},
		"name": {
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			Description:  "Cluster name. Changing it recreates the cluster.",
			ValidateFunc: validateKubernetesResourceName,
		},
		"version": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "Kubernetes version.",
		},
		"location": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Cluster location.",
		},
		"needs_service_vms": {
			Type:        schema.TypeBool,
			Description: "Deploy Service VMs providing load balancers and outbound masquerade.",
			Default:     true,
			Optional:    true,
		},
		"enable_nat_gateways": {
			Type:        schema.TypeBool,
			Description: "If enabled, Service VMs are configured as NAT gateways connecting the internal cluster network to the internet.",
			Default:     true,
			Optional:    true,
		},
		"enable_lbaas": {
			Type:        schema.TypeBool,
			Description: "If enabled, Service VMs are set up as LBaaS hosts enabling K8s services of type LoadBalancer.",
			Default:     true,
			Optional:    true,
		},
		"internal_ipv4_prefix": {
			Type:        schema.TypeString,
			Description: "Internal IPv4 network prefix identifier. Use the `id` exported by an `anxcloud_network_prefix` resource or data source.",
			Optional:    true,
			Computed:    true,
		},
		"external_ipv4_prefix": {
			Type:        schema.TypeString,
			Description: "External IPv4 network prefix identifier. Use the `id` exported by an `anxcloud_network_prefix` resource or data source.",
			Optional:    true,
			Computed:    true,
		},
		"external_ipv6_prefix": {
			Type:        schema.TypeString,
			Description: "External IPv6 network prefix identifier. Use the `id` exported by an `anxcloud_network_prefix` resource or data source.",
			Optional:    true,
			Computed:    true,
		},
		"enable_autoscaling": {
			Type: schema.TypeBool,
			Description: `
Enable autoscaling for this cluster. Defaults to false if unset.

-> You will need to explicitly configure your node pools for autoscaling. Please check the provided [autoscaling documentation](https://engine.anexia-it.com/docs/en/module/kubernetes/user-guide/autoscaling) for details.`,
			Optional: true,
		},
		"apiserver_allowlist": {
			Type:        schema.TypeList,
			Description: "A list of CIDRs that should be allowed access to the kubernetes API server. By default there are no IP restrictions.",
			Optional:    true,
			Computed:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"cni_plugin": {
			Type:         schema.TypeString,
			Optional:     true,
			Computed:     true,
			Description:  "Container Network Interface plugin. Currently only Canal is supported.",
			ValidateFunc: validation.StringInSlice([]string{"canal"}, false),
		},
		"external_ip_families": {
			Type:         schema.TypeString,
			Optional:     true,
			Computed:     true,
			Description:  "IP families used for external networking. Valid values are `IPv4` and `Dualstack`.",
			ValidateFunc: validation.StringInSlice([]string{"IPv4", "Dualstack"}, false),
		},
		"enable_oidc_authentication": {
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
			Description: "Enable OIDC authentication for the Kubernetes cluster.",
		},
		"oidc_client_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "OIDC client ID.",
		},
		"oidc_issuer_url": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "OIDC issuer URL.",
		},
		"oidc_groups_claim": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "OIDC claim used to determine user groups.",
		},
		"oidc_username_claim": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "OIDC claim used to determine the username.",
		},
		"oidc_extra_scopes": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Space-separated list of additional OIDC scopes.",
		},
		"oidc_groups_prefix": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Prefix applied when filtering OIDC group claims.",
		},
		"oidc_required_claim": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "OIDC claim a user must have.",
		},
		"oidc_username_prefix": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Prefix applied when filtering OIDC usernames.",
		},
		"maintenance_window_start_time": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Maintenance window start in UTC, for example `Tue 22:00` or `22:00`.",
		},
		"maintenance_window_duration": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Maintenance window duration, for example `2h`, `30m`, or `15h30m`.",
		},
		"patch_version": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Current Kubernetes patch version.",
		},
		"state": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Current reconciliation state identifier.",
		},
		"state_text": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Human-readable current reconciliation state.",
		},
	}
}

func schemaKubernetesNodePool() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"api_environment": schemaKubernetesAPIEnvironment(),
		"id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Node pool identifier.",
		},
		"name": {
			Type:         schema.TypeString,
			Required:     true,
			Description:  "Node pool name.",
			ValidateFunc: validateKubernetesResourceName,
		},
		"initial_replicas": {
			Type:         schema.TypeInt,
			Required:     true,
			Description:  "Desired number of nodes. When autoscaling is enabled, the API reports the current replica count without overwriting this configured value.",
			ValidateFunc: validation.IntBetween(0, 100),
		},
		"cpus": {
			Type:         schema.TypeInt,
			Required:     true,
			Description:  "Number of CPUs per node.",
			ValidateFunc: validation.IntBetween(1, 16),
		},
		"memory_gib": {
			Type:         schema.TypeInt,
			Required:     true,
			Description:  "Memory per node in GiB.",
			ValidateFunc: validation.IntBetween(2, 64),
		},
		"disk": {
			Required:    true,
			MinItems:    1,
			MaxItems:    1,
			Description: "List of disks for each node.",
			Type:        schema.TypeList,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"size_gib": {
						Type:         schema.TypeInt,
						Required:     true,
						Description:  "Disk size in GiB.",
						ValidateFunc: validation.IntBetween(20, 1600),
					},
					"performance_type": {
						Type:         schema.TypeString,
						Optional:     true,
						Computed:     true,
						Description:  "Disk performance type.",
						ValidateFunc: validateKubernetesDiskPerformanceType,
					},
				},
			},
		},
		"operating_system": {
			Type:         schema.TypeString,
			Required:     true,
			Description:  `Operating system. Only "Flatcar Linux" supported at the moment.`,
			ValidateFunc: validation.StringInSlice([]string{"Flatcar Linux"}, false),
		},
		"cluster": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Cluster identifier.",
		},
		"sync_source": {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "engine",
			Description:  "Source of truth for node pool configuration. Valid values are `engine` and `cluster`; defaults to `engine`.",
			ValidateFunc: validation.StringInSlice([]string{"engine", "cluster"}, true),
			StateFunc: func(value any) string {
				return strings.ToLower(value.(string))
			},
		},
		"cpu_performance_type": {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "performance",
			Description:  "CPU performance type.",
			ValidateFunc: validation.StringInSlice([]string{"best-effort", "standard", "enterprise", "performance", "performance-plus"}, false),
		},
		"autoscaler_enabled": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Enable automatic node count adjustment.",
		},
		"autoscaler_min_nodes": {
			Type:         schema.TypeInt,
			Optional:     true,
			Default:      0,
			Description:  "Minimum node count used by the autoscaler.",
			ValidateFunc: validation.IntBetween(0, 100),
		},
		"autoscaler_max_nodes": {
			Type:         schema.TypeInt,
			Optional:     true,
			Default:      0,
			Description:  "Maximum node count used by the autoscaler.",
			ValidateFunc: validation.IntBetween(0, 100),
		},
		"additional_disks": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    10,
			Description: "Additional disks attached to each node.",
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"name": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "Disk name.",
				},
				"size_gib": {
					Type:         schema.TypeInt,
					Required:     true,
					Description:  "Disk size in GiB.",
					ValidateFunc: validation.IntBetween(20, 1600),
				},
				"performance_type": {
					Type:         schema.TypeString,
					Required:     true,
					Description:  "Disk performance type.",
					ValidateFunc: validateKubernetesDiskPerformanceType,
				},
			}},
		},
		"networks": {
			Type:        schema.TypeList,
			Required:    true,
			MinItems:    1,
			MaxItems:    10,
			Description: "Network interfaces attached to each node. The API requires between one and ten networks during node-pool creation.",
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"name": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "Network interface name.",
				},
				"bandwidth_limit": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "Network bandwidth limit identifier.",
				},
				"vlan": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "VLAN identifier.",
				},
			}},
		},
		"dns_override_ipv4": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Enable custom IPv4 DNS servers.",
		},
		"dns_ipv4_1": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "First IPv4 DNS server.",
		},
		"dns_ipv4_2": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Second IPv4 DNS server.",
		},
		"dns_override_ipv6": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Enable custom IPv6 DNS servers.",
		},
		"dns_ipv6_1": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "First IPv6 DNS server.",
		},
		"dns_ipv6_2": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Second IPv6 DNS server.",
		},
		"ssh_public_keys": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "SSH public keys in authorized_keys format, separated by line breaks.",
		},
		"taints": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Kubernetes taints separated by line breaks.",
		},
		"labels": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Kubernetes labels separated by line breaks.",
		},
		"annotations": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Kubernetes annotations separated by line breaks.",
		},
		"state": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Current reconciliation state identifier.",
		},
		"state_text": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Human-readable current reconciliation state.",
		},
	}
}

var validateKubernetesDiskPerformanceType = validation.StringInSlice([]string{
	"STD1", "STD2", "STD3", "STD4", "STD5",
	"ENT1", "ENT2", "ENT3", "ENT4", "ENT5", "ENT6",
}, false)

func schemaKubernetesKubeConfig() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"api_environment": schemaKubernetesAPIEnvironment(),
		"cluster": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Cluster identifier.",
			ForceNew:    true,
		},

		"host": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Cluster control-plane host.",
		},
		"token": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Kubeconfig token.",
			Sensitive:   true,
		},
		"cluster_ca_certificate": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Kubeconfig cluster ca certificate.",
			Sensitive:   true,
		},

		"raw": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Raw kubeconfig.",
			Sensitive:   true,
		},
	}
}

func schemaKubernetesAPIEnvironment() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		Default:     kubernetesAPIEnvironmentProd,
		ForceNew:    true,
		Description: "Kubernetes service environment. Valid values are `prod`, `stage`, and `dev`; defaults to `prod`. Use the same value for a cluster and its node pools and kubeconfigs. For managed resources, changing it recreates the resource because it selects a different API endpoint.",
		ValidateFunc: validation.StringInSlice([]string{
			kubernetesAPIEnvironmentProd,
			kubernetesAPIEnvironmentStage,
			kubernetesAPIEnvironmentDev,
		}, false),
	}
}
