package anxcloud

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"runtime"

	"github.com/go-logr/logr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"go.anx.io/go-anxcloud/pkg/api"
	"go.anx.io/go-anxcloud/pkg/client"
)

var providerVersion = "development"

var logger logr.Logger

// Provider Anexia
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"token": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("ANEXIA_TOKEN", nil),
				Description: "Anexia Cloud token.",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"anxcloud_virtual_server":        resourceVirtualServer(),
			"anxcloud_vlan":                  resourceVLAN(),
			"anxcloud_network_prefix":        resourceNetworkPrefix(),
			"anxcloud_ip_address":            resourceIPAddress(),
			"anxcloud_tag":                   resourceTag(),
			"anxcloud_dns_zone":              resourceDNSZone(),
			"anxcloud_dns_record":            resourceDNSRecord(),
			"anxcloud_lbaas_loadbalancer":    resourceLBaaSLoadBalancer(),
			"anxcloud_kubernetes_cluster":    resourceKubernetesCluster(),
			"anxcloud_kubernetes_node_pool":  resourceKubernetesNodePool(),
			"anxcloud_kubernetes_kubeconfig": resourceKubernetesKubeconfig(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"anxcloud_disk_types":            dataSourceDiskTypes(),
			"anxcloud_template":              dataSourceTemplate(),
			"anxcloud_ip_address":            dataSourceIPAddress(),
			"anxcloud_ip_addresses":          dataSourceIPAddresses(),
			"anxcloud_nic_types":             dataSourceNICTypes(),
			"anxcloud_core_location":         dataSourceCoreLocation(),
			"anxcloud_core_locations":        dataSourceCoreLocations(),
			"anxcloud_vlans":                 dataSourceVLANs(),
			"anxcloud_tags":                  dataSourceTags(),
			"anxcloud_cpu_performance_types": dataSourceCPUPerformanceTypes(),
			"anxcloud_dns_records":           dataSourceDNSRecords(),
			"anxcloud_dns_zones":             datasourceDNSZones(),
			"anxcloud_kubernetes_cluster":    dataSourceKubernetesCluster(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

type providerContext struct {
	api          api.API
	legacyClient client.Client
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	logger = NewTerraformr(log.Default().Writer())
	var diags diag.Diagnostics

	token := d.Get("token").(string)

	opts := []client.Option{
		client.TokenFromString(token),
		client.Logger(logger.WithName("client")),
		client.UserAgent(fmt.Sprintf("%s/%s (%s)", "terraform-provider-anxcloud", providerVersion, runtime.GOOS)),
	}

	c, err := client.New(opts...)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create Anexia client",
			Detail:   "Unable to create Anexia client with the given token, either the token is empty or invalid",
		})
		return nil, diags
	}

	apiClient, err := api.NewAPI(api.WithClientOptions(opts...))
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create generic Anexia client",
			Detail:   "Unable to create generic Anexia client with the given token, either the token is empty or invalid",
		})
		return nil, diags
	}

	return providerContext{
		api:          apiClient,
		legacyClient: c,
	}, diags
}

func handleNotFoundError(err error) error {
	var respErr *client.ResponseError
	if errors.As(err, &respErr) && respErr.ErrorData.Code == http.StatusNotFound {
		return nil
	}
	return err
}

// context key type for provider package
type providerContextKey string

func apiFromProviderConfig(m interface{}) api.API {
	return m.(providerContext).api
}
