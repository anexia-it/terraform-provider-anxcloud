package anxcloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"go.anx.io/go-anxcloud/pkg/api"
	frontierv1 "go.anx.io/go-anxcloud/pkg/apis/frontier/v1"
)

func resourceFrontierDeployment() *schema.Resource {
	return &schema.Resource{
		Description:   "A deployment represents a published version of a Frontier API with all its endpoints and actions exactly as it was at the time it was deployed.",
		CreateContext: resourceFrontierDeploymentCreate,
		ReadContext:   resourceFrontierDeploymentRead,
		DeleteContext: resourceFrontierDeploymentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Deployment identifier.",
			},
			"slug": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Deployment slug.",
			},
			"api": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Deployment API identifier.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Deployment name.",
			},
			"state": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Deployment state.",
			},
			"revision": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "Deployment revision is an optional attribute which can be used to trigger a new deployment." +
					" Use the `create_before_destroy` lifecycle argument to ensure that there is always a deployment present." +
					" The value can be any arbitrary string (e.g. `COMMIT_SHA` passed in via variables).",
			},
		},
	}
}

func resourceFrontierDeploymentCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	a := apiFromProviderConfig(m)

	frontierDeployment := frontierv1.Deployment{
		APIIdentifier: d.Get("api").(string),
		Slug:          d.Get("slug").(string),
	}

	if err := a.Create(ctx, &frontierDeployment); err != nil {
		return diag.Errorf("failed to create resource: %s", err)
	}

	d.SetId(frontierDeployment.Identifier)

	err := retry.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *retry.RetryError {
		if err := a.Get(ctx, &frontierDeployment); err != nil {
			return retry.NonRetryableError(err)
		}

		if frontierDeployment.State == "deploying" {
			return retry.RetryableError(fmt.Errorf("resource still pending"))
		}

		if frontierDeployment.State != "deployed" {
			return retry.NonRetryableError(fmt.Errorf("unexpected state: %q", frontierDeployment.State))
		}

		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceFrontierDeploymentRead(ctx, d, m)
}

func resourceFrontierDeploymentRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	a := apiFromProviderConfig(m)

	frontierDeployment := frontierv1.Deployment{Identifier: d.Id()}
	if err := a.Get(ctx, &frontierDeployment); api.IgnoreNotFound(err) != nil {
		return diag.Errorf("failed getting resource: %s", err)
	} else if err != nil {
		d.SetId("")
		return nil
	}

	var diags diag.Diagnostics

	if err := d.Set("name", frontierDeployment.Name); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("slug", frontierDeployment.Slug); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("api", frontierDeployment.APIIdentifier); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("state", frontierDeployment.State); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	return diags
}

func resourceFrontierDeploymentDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	a := apiFromProviderConfig(m)

	if err := a.Destroy(ctx, &frontierv1.Deployment{Identifier: d.Id()}); api.IgnoreNotFound(err) != nil {
		return diag.Errorf("failed deleting resource: %s", err)
	}

	return diag.FromErr(retry.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *retry.RetryError {
		if err := a.Get(ctx, &frontierv1.Deployment{Identifier: d.Id()}); api.IgnoreNotFound(err) != nil {
			return retry.NonRetryableError(err)
		} else if err != nil {
			return nil
		}
		return retry.RetryableError(fmt.Errorf("resource still deleting"))
	}))
}
