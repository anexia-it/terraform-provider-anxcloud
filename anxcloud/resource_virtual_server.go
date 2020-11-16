package anxcloud

import (
	"context"
	"fmt"
	"time"

	"github.com/anexia-it/go-anxcloud/pkg/client"
	"github.com/anexia-it/go-anxcloud/pkg/vsphere"
	"github.com/anexia-it/go-anxcloud/pkg/vsphere/provisioning/vm"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceVirtualServer() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVirtualServerCreate,
		ReadContext:   resourceVirtualServerRead,
		UpdateContext: resourceVirtualServerUpdate,
		DeleteContext: resourceVirtualServerDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Read:   schema.DefaultTimeout(1 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Virtual server identifier.",
			},
			"location_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Location identifier.",
			},
			"template_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Template identifier.",
			},
			"template_type": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "OS template type.",
			},
			"hostname": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Virtual server hostname.",
			},
			"cpus": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Amount of CPUs.",
			},
			"memory": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Memory in MB.",
			},
			"disk": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Disk capacity in GB.",
			},
			"network": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Network interfaces",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"vlan_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "VLAN identifier.",
						},
						"nic_type": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Network interface card type.",
						},
						"ips": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     schema.TypeString,
							Description: "List of IPs and IPs identifiers. IPs are ignored when using template_type 'from_scratch'." +
								"Defaults to free IPs from IP pool attached to VLAN.",
						},
					},
				},
			},
		},
	}
}

func resourceVirtualServerCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var (
		diags    diag.Diagnostics
		networks []vm.Network
	)

	locationID := d.Get("location_id").(string)
	c := m.(client.Client)
	v := vsphere.NewAPI(c)

	networks = expandNetworks(d.Get("networks").([]interface{}))
	for _, n := range networks {
		var ips []string

		if len(n.IPs) > 0 {
			ips = n.IPs
		} else {
			freeIPs, err := v.Provisioning().IPs().GetFree(ctx, locationID, n.VLAN)
			if err != nil {
				diags = append(diags, diag.FromErr(err)...)
			}
			ip := freeIPs[0]
			ips = append(ips, ip.Identifier)
		}

		networks = append(networks, vm.Network{
			NICType: n.NICType,
			VLAN:    n.VLAN,
			IPs:     ips,
		})
	}

	if len(diags) > 0 {
		return diags
	}

	def := v.Provisioning().VM().NewDefinition(
		locationID,
		d.Get("template_type").(string),
		d.Get("template_id").(string),
		d.Get("hostname").(string),
		d.Get("cpus").(int),
		d.Get("memory").(int),
		d.Get("disk").(int),
		networks,
	)

	provision, err := v.Provisioning().VM().Provision(ctx, def)
	if err != nil {
		return diag.FromErr(err)
	}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		const complete = 100

		p, err := v.Provisioning().Progress().Get(ctx, provision.Identifier)
		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("unable to get VM progress by ID '%s', %w", provision.Identifier, err))
		}
		if p.Progress == complete && p.VMIdentifier != "" {
			d.SetId(p.VMIdentifier)
			return nil
		}
		return resource.RetryableError(fmt.Errorf("VM with provisioning id '%s' is not ready yet: %d %%", provision.Identifier, p.Progress))
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return resourceVirtualServerRead(ctx, d, m)
}

func resourceVirtualServerRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	return diags
}

func resourceVirtualServerUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	if len(diags) > 0 {
		return diags
	}

	return resourceVirtualServerRead(ctx, d, m)
}

func resourceVirtualServerDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	return diags
}
