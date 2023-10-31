package anxcloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"go.anx.io/go-anxcloud/pkg/api"
	"go.anx.io/go-anxcloud/pkg/utils/pointer"

	frontierv1 "go.anx.io/go-anxcloud/pkg/apis/frontier/v1"
)

func resourceFrontierAPI() *schema.Resource {
	return &schema.Resource{
		Description:   "An API represents Frontier's root object and contains a collection of endpoints. The API defines the transfer protocol, such as HTTP and HTTPS, for all containing endpoints.",
		CreateContext: resourceFrontierAPICreate,
		ReadContext:   resourceFrontierAPIRead,
		UpdateContext: resourceFrontierAPIUpdate,
		DeleteContext: resourceFrontierAPIDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{

			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "API identifier.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "API name.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "API description.",
			},
			"transfer_protocol": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"http"}, false),
				Description:  "API transfer protocol. Currently `http` is the only supported value.",
			},
		},
	}
}

func resourceFrontierAPICreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	a := apiFromProviderConfig(m)

	frontierAPI := frontierv1.API{
		Name:             d.Get("name").(string),
		Description:      pointer.String(d.Get("description").(string)),
		TransferProtocol: d.Get("transfer_protocol").(string),
	}

	if err := a.Create(ctx, &frontierAPI); err != nil {
		return diag.Errorf("failed to create resource: %s", err)
	}

	d.SetId(frontierAPI.Identifier)

	return resourceFrontierAPIRead(ctx, d, m)
}

func resourceFrontierAPIRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	a := apiFromProviderConfig(m)

	frontierAPI := frontierv1.API{Identifier: d.Id()}
	if err := a.Get(ctx, &frontierAPI); api.IgnoreNotFound(err) != nil {
		return diag.Errorf("failed getting resource: %s", err)
	} else if err != nil {
		d.SetId("")
		return nil
	}

	var diags diag.Diagnostics

	if err := d.Set("name", frontierAPI.Name); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("description", pointer.StringVal(frontierAPI.Description)); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("transfer_protocol", frontierAPI.TransferProtocol); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	return diags
}

func resourceFrontierAPIUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	a := apiFromProviderConfig(m)

	frontierAPI := frontierv1.API{
		Identifier:       d.Id(),
		Name:             d.Get("name").(string),
		Description:      pointer.String(d.Get("description").(string)),
		TransferProtocol: d.Get("transfer_protocol").(string),
	}

	if err := a.Update(ctx, &frontierAPI); err != nil {
		return diag.Errorf("failed updating resource: %s", err)
	}

	return resourceFrontierAPIRead(ctx, d, m)
}

func resourceFrontierAPIDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	a := apiFromProviderConfig(m)

	if err := a.Destroy(ctx, &frontierv1.API{Identifier: d.Id()}); api.IgnoreNotFound(err) != nil {
		return diag.Errorf("failed deleting resource: %s", err)
	}

	return nil
}
