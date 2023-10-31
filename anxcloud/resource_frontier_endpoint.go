package anxcloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"go.anx.io/go-anxcloud/pkg/api"
	frontierv1 "go.anx.io/go-anxcloud/pkg/apis/frontier/v1"
)

func resourceFrontierEndpoint() *schema.Resource {
	return &schema.Resource{
		Description:   "An endpoint represents a path within an HTTP-based API and contains a collection of actions.",
		CreateContext: resourceFrontierEndpointCreate,
		ReadContext:   resourceFrontierEndpointRead,
		UpdateContext: resourceFrontierEndpointUpdate,
		DeleteContext: resourceFrontierEndpointDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Endpoint identifier.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Endpoint name.",
			},
			"path": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Endpoint path.",
			},
			"api": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Endpoint API identifier.",
			},
		},
	}
}

func resourceFrontierEndpointCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	a := apiFromProviderConfig(m)

	frontierEndpoint := frontierv1.Endpoint{
		Name:          d.Get("name").(string),
		Path:          d.Get("path").(string),
		APIIdentifier: d.Get("api").(string),
	}

	if err := a.Create(ctx, &frontierEndpoint); err != nil {
		return diag.Errorf("failed to create resource: %s", err)
	}

	d.SetId(frontierEndpoint.Identifier)

	return resourceFrontierEndpointRead(ctx, d, m)
}

func resourceFrontierEndpointRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	a := apiFromProviderConfig(m)

	frontierEndpoint := frontierv1.Endpoint{Identifier: d.Id()}
	if err := a.Get(ctx, &frontierEndpoint); api.IgnoreNotFound(err) != nil {
		return diag.Errorf("failed getting resource: %s", err)
	} else if err != nil {
		d.SetId("")
		return nil
	}

	var diags diag.Diagnostics

	if err := d.Set("name", frontierEndpoint.Name); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("path", frontierEndpoint.Path); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("api", frontierEndpoint.APIIdentifier); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	return diags
}

func resourceFrontierEndpointUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	a := apiFromProviderConfig(m)

	frontierEndpoint := frontierv1.Endpoint{
		Identifier:    d.Id(),
		Name:          d.Get("name").(string),
		Path:          d.Get("path").(string),
		APIIdentifier: d.Get("api").(string),
	}

	if err := a.Update(ctx, &frontierEndpoint); err != nil {
		return diag.Errorf("failed updating resource: %s", err)
	}

	return resourceFrontierEndpointRead(ctx, d, m)
}

func resourceFrontierEndpointDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	a := apiFromProviderConfig(m)

	if err := a.Destroy(ctx, &frontierv1.Endpoint{Identifier: d.Id()}); api.IgnoreNotFound(err) != nil {
		return diag.Errorf("failed deleting resource: %s", err)
	}

	return nil
}
