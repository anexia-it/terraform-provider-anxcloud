package anxcloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func schemaKubernetesCluster() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Cluster identifier.",
		},
		"name": {
			Type:         schema.TypeString,
			Required:     true,
			Description:  "Cluster name.",
			ForceNew:     true,
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
			ForceNew:    true,
		},
		"needs_service_vms": {
			Type:        schema.TypeBool,
			Description: "Deploy Service VMs providing load balancers and outbound masquerade.",
			ForceNew:    true,
			Default:     true,
			Optional:    true,
		},
		"enable_nat_gateways": {
			Type:        schema.TypeBool,
			Description: "If enabled, Service VMs are configured as NAT gateways connecting the internal cluster network to the internet.",
			ForceNew:    true,
			Default:     true,
			Optional:    true,
		},
		"enable_lbaas": {
			Type:        schema.TypeBool,
			Description: "If enabled, Service VMs are set up as LBaaS hosts enabling K8s services of type LoadBalancer.",
			ForceNew:    true,
			Default:     true,
			Optional:    true,
		},
		"internal_ipv4_prefix": {
			Type:        schema.TypeString,
			Description: "Internal IPv4 prefix.",
			Optional:    true,
			Computed:    true,
			ForceNew:    true,
		},
		"external_ipv4_prefix": {
			Type:        schema.TypeString,
			Description: "External IPv4 prefix.",
			Optional:    true,
			Computed:    true,
			ForceNew:    true,
		},
		"external_ipv6_prefix": {
			Type:        schema.TypeString,
			Description: "External IPv6 prefix.",
			Optional:    true,
			Computed:    true,
			ForceNew:    true,
		},
		"enable_autoscaling": {
			Type: schema.TypeBool,
			Description: `
Enable autoscaling for this cluster. Defaults to false if unset.

-> You will need to explicitly configure your node pools for autoscaling. Please check the provided [autoscaling documentation](https://engine.anexia-it.com/docs/en/module/kubernetes/user-guide/autoscaling) for details.`,
			Optional: true,
			ForceNew: true,
		},
	}
}

func schemaKubernetesNodePool() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Node pool identifier.",
		},
		"name": {
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			Description:  "Node pool name.",
			ValidateFunc: validateKubernetesResourceName,
		},
		"initial_replicas": {
			Type:        schema.TypeInt,
			Required:    true,
			ForceNew:    true,
			Description: "Initial number of nodes.",
		},
		"cpus": {
			Type:        schema.TypeInt,
			Required:    true,
			ForceNew:    true,
			Description: "Number of CPUs per node.",
		},
		"memory_gib": {
			Type:        schema.TypeInt,
			Required:    true,
			ForceNew:    true,
			Description: "Memory per node in GiB.",
		},
		"disk": {
			Required:    true,
			ForceNew:    true,
			MinItems:    1,
			MaxItems:    1,
			Description: "List of disks for each node.",
			Type:        schema.TypeList,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"size_gib": {
						Type:        schema.TypeInt,
						Required:    true,
						ForceNew:    true,
						Description: "Disk size in GiB.",
					},
				},
			},
		},
		"operating_system": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: `Operating system. Only "Flatcar Linux" supported at the moment.`,
		},
		"cluster": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "Cluster identifier.",
		},
	}
}

func schemaKubernetesKubeConfig() map[string]*schema.Schema {
	return map[string]*schema.Schema{
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
