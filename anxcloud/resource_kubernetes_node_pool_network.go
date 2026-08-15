package anxcloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"go.anx.io/go-anxcloud/pkg/api"
)

func resourceKubernetesNodePoolNetwork() *schema.Resource {
	return &schema.Resource{
		Description: "Resource to attach and update an additional VLAN network on a Kubernetes node pool.",

		CreateContext: resourceKubernetesNodePoolNetworkCreate,
		ReadContext:   resourceKubernetesNodePoolNetworkRead,
		UpdateContext: resourceKubernetesNodePoolNetworkUpdate,
		DeleteContext: resourceKubernetesNodePoolNetworkDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Read:   schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: schemaKubernetesNodePoolNetwork(),
	}
}

func resourceKubernetesNodePoolNetworkCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	a := apiFromProviderConfig(m)
	network := kubernetesNodePoolNetwork{requestDefinition: kubernetesNodePoolNetworkCreateDefinition(d)}

	if err := a.Create(ctx, &network); err != nil {
		return diag.Errorf("failed to create Kubernetes node-pool network: %s", err)
	}

	d.SetId(network.Identifier)
	return setResourceDataFromKubernetesNodePoolNetwork(d, network)
}

func resourceKubernetesNodePoolNetworkRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	network, err := getKubernetesNodePoolNetwork(ctx, apiFromProviderConfig(m), d.Id())
	if err != nil {
		if api.IgnoreNotFound(err) == nil {
			d.SetId("")
			return nil
		}
		return diag.Errorf("failed getting Kubernetes node-pool network: %s", err)
	}

	return setResourceDataFromKubernetesNodePoolNetwork(d, network)
}

func resourceKubernetesNodePoolNetworkUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	definition := kubernetesNodePoolNetworkUpdateDefinition(d)
	if len(definition) == 0 {
		return resourceKubernetesNodePoolNetworkRead(ctx, d, m)
	}

	network := kubernetesNodePoolNetwork{requestDefinition: definition}
	network.Identifier = d.Id()
	if err := apiFromProviderConfig(m).Update(ctx, &network); err != nil {
		return diag.Errorf("failed to update Kubernetes node-pool network: %s", err)
	}

	return setResourceDataFromKubernetesNodePoolNetwork(d, network)
}

func resourceKubernetesNodePoolNetworkDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	network := kubernetesNodePoolNetwork{}
	network.Identifier = d.Id()
	if err := apiFromProviderConfig(m).Destroy(ctx, &network); err != nil && api.IgnoreNotFound(err) != nil {
		return diag.Errorf("failed deleting Kubernetes node-pool network: %s", err)
	}

	d.SetId("")
	return nil
}

func kubernetesNodePoolNetworkCreateDefinition(d *schema.ResourceData) map[string]any {
	definition := map[string]any{
		"name":            d.Get("name").(string),
		"nodepool":        d.Get("node_pool").(string),
		"bandwidth_limit": d.Get("bandwidth_limit").(string),
		"vlan":            d.Get("vlan").(string),
	}
	return definition
}

func kubernetesNodePoolNetworkUpdateDefinition(d *schema.ResourceData) map[string]any {
	definition := make(map[string]any)
	fields := map[string]string{
		"name":            "name",
		"node_pool":       "nodepool",
		"bandwidth_limit": "bandwidth_limit",
		"vlan":            "vlan",
	}
	for terraformName, apiName := range fields {
		if d.HasChange(terraformName) {
			definition[apiName] = d.Get(terraformName)
		}
	}
	return definition
}

func getKubernetesNodePoolNetwork(ctx context.Context, a api.API, identifier string) (kubernetesNodePoolNetwork, error) {
	network := kubernetesNodePoolNetwork{}
	network.Identifier = identifier
	if err := a.Get(ctx, &network); err != nil {
		return network, err
	}
	return network, nil
}

func setResourceDataFromKubernetesNodePoolNetwork(d *schema.ResourceData, network kubernetesNodePoolNetwork) diag.Diagnostics {
	var diags diag.Diagnostics
	set := func(name string, value any) {
		if err := d.Set(name, value); err != nil {
			diags = append(diags, diag.FromErr(fmt.Errorf("set %s: %w", name, err))...)
		}
	}

	set("name", network.Name)
	set("node_pool", network.NodePool.Identifier)
	set("bandwidth_limit", network.BandwidthLimit.ID)
	set("vlan", network.VLAN.Identifier)
	set("state", network.State.ID)
	set("state_text", network.State.Text)
	return diags
}
