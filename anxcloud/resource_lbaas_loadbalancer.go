package anxcloud

import (
	"context"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"go.anx.io/go-anxcloud/pkg/api"
	lbaasv1 "go.anx.io/go-anxcloud/pkg/apis/lbaas/v1"
)

func resourceLBaaSLoadBalancer() *schema.Resource {
	return &schema.Resource{
		Description: strings.TrimSpace(`
This resource allows you to create and manage LBaaS Load Balancer resources.

-> This process is not yet fully automated, we are working on it.
`),
		CreateContext: tagsMiddlewareCreate(resourceLBaaSLoadBalancerCreate),
		ReadContext:   tagsMiddlewareRead(resourceLBaaSLoadBalancerRead),
		UpdateContext: tagsMiddlewareUpdate(resourceLBaaSLoadBalancerUpdate),
		DeleteContext: resourceLBaaSLoadBalancerDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Second),
			Read:   schema.DefaultTimeout(30 * time.Second),
			Update: schema.DefaultTimeout(30 * time.Second),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: withTagsAttribute(schemaLBaaSLoadBalancer()),
	}
}

func resourceLBaaSLoadBalancerCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	a := m.(providerContext).api

	loadBalancer := lbaasv1.LoadBalancer{
		Name:      d.Get("name").(string),
		IpAddress: d.Get("ip_address").(string),
	}

	if err := a.Create(ctx, &loadBalancer); err != nil {
		return diag.Errorf("failed to create LoadBalancer: %s", err)
	}

	d.SetId(loadBalancer.Identifier)

	return resourceLBaaSLoadBalancerRead(ctx, d, m)
}

func resourceLBaaSLoadBalancerRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	a := m.(providerContext).api

	loadBalancer := lbaasv1.LoadBalancer{Identifier: d.Id()}

	if err := a.Get(ctx, &loadBalancer); err != nil {
		return diag.Errorf("failed to get LoadBalancer: %s", err)
	}

	var diags diag.Diagnostics

	if err := d.Set("name", loadBalancer.Name); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("ip_address", loadBalancer.IpAddress); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	return diags
}

func resourceLBaaSLoadBalancerUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	a := m.(providerContext).api

	loadBalancer := lbaasv1.LoadBalancer{
		Identifier: d.Id(),
		Name:       d.Get("name").(string),
		IpAddress:  d.Get("ip_address").(string),
	}

	if err := a.Update(ctx, &loadBalancer); err != nil {
		return diag.Errorf("failed to update LoadBalancer: %s", err)
	}

	return resourceLBaaSLoadBalancerRead(ctx, d, m)
}

func resourceLBaaSLoadBalancerDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	a := m.(providerContext).api

	loadBalancer := lbaasv1.LoadBalancer{Identifier: d.Id()}

	if err := a.Destroy(ctx, &loadBalancer); api.IgnoreNotFound(err) != nil {
		return diag.Errorf("failed to delete LoadBalancer: %s", err)
	}

	d.SetId("")

	return nil
}
