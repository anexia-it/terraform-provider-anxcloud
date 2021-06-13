package anxcloud

import (
	"context"
	"errors"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	"log"
	"net/http"

	"github.com/anexia-it/go-anxcloud/pkg/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

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
			"anxcloud_virtual_server": resourceVirtualServer(),
			"anxcloud_vlan":           resourceVLAN(),
			"anxcloud_network_prefix": resourceNetworkPrefix(),
			"anxcloud_ip_address":     resourceIPAddress(),
			"anxcloud_tag":            resourceTag(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"anxcloud_disk_types":            dataSourceDiskTypes(),
			"anxcloud_template":              dataSourceTemplate(),
			"anxcloud_ip_addresses":          dataSourceIPAddresses(),
			"anxcloud_nic_types":             dataSourceNICTypes(),
			"anxcloud_core_locations":        dataSourceCoreLocations(),
			"anxcloud_vlans":                 dataSourceVLANs(),
			"anxcloud_tags":                  dataSourceTags(),
			"anxcloud_cpu_performance_types": dataSourceCPUPerformanceTypes(),
			"anxcloud_vsphere_locations":     dataSourceVSphereLocations(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	token := d.Get("token").(string)
	opts := []client.Option{
		client.TokenFromString(token),
	}

	debugLogWriter := debugWriter{
		writer: log.Writer(),
	}
	if logging.LogLevel() != "" {
		logOpt := client.LogWriter(debugLogWriter)
		opts = append(opts, logOpt)
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

	return c, diags
}

func handleNotFoundError(err error) error {
	var respErr *client.ResponseError
	if errors.As(err, &respErr) && respErr.ErrorData.Code == http.StatusNotFound {
		return nil
	}
	return err
}
