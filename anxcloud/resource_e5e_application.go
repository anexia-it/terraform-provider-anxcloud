package anxcloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"go.anx.io/go-anxcloud/pkg/api"

	e5ev1 "go.anx.io/go-anxcloud/pkg/apis/e5e/v1"
)

func resourceE5EApplication() *schema.Resource {
	return &schema.Resource{
		Description: "Applications are an easy way to bring more structure to your configured functions by grouping them." +
			" You can imagine an application as a folder to put in your functions.",
		CreateContext: resourceE5EApplicationCreate,
		ReadContext:   resourceE5EApplicationRead,
		UpdateContext: resourceE5EApplicationUpdate,
		DeleteContext: resourceE5EApplicationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Application identifier.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Application name.",
			},
		},
	}
}

func resourceE5EApplicationCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	a := apiFromProviderConfig(m)

	application := e5ev1.Application{
		Name: d.Get("name").(string),
	}

	if err := a.Create(ctx, &application); err != nil {
		return diag.Errorf("failed to create resource: %s", err)
	}

	d.SetId(application.Identifier)

	return resourceE5EApplicationRead(ctx, d, m)
}

func resourceE5EApplicationRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	a := apiFromProviderConfig(m)

	application := e5ev1.Application{Identifier: d.Id()}
	if err := a.Get(ctx, &application); api.IgnoreNotFound(err) != nil {
		return diag.Errorf("failed getting resource: %s", err)
	} else if err != nil {
		d.SetId("")
		return nil
	}

	var diags diag.Diagnostics

	if err := d.Set("name", application.Name); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	return diags
}

func resourceE5EApplicationUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	a := apiFromProviderConfig(m)

	application := e5ev1.Application{
		Identifier: d.Id(),
		Name:       d.Get("name").(string),
	}

	if err := a.Update(ctx, &application); err != nil {
		return diag.Errorf("failed updating resource: %s", err)
	}

	return resourceE5EApplicationRead(ctx, d, m)
}

func resourceE5EApplicationDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	a := apiFromProviderConfig(m)

	if err := a.Destroy(ctx, &e5ev1.Application{Identifier: d.Id()}); api.IgnoreNotFound(err) != nil {
		return diag.Errorf("failed deleting resource %s", err)
	}

	return nil
}
