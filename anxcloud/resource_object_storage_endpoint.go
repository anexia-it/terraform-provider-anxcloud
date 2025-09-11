package anxcloud

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"go.anx.io/go-anxcloud/pkg/api"
	objectstoragev2 "go.anx.io/go-anxcloud/pkg/apis/objectstorage/v2"
)

func resourceObjectStorageEndpoint() *schema.Resource {
	return &schema.Resource{
		Description:   "This resource allows you to create and manage Object Storage endpoints.",
		CreateContext: resourceObjectStorageEndpointCreate,
		ReadContext:   resourceObjectStorageEndpointRead,
		UpdateContext: resourceObjectStorageEndpointUpdate,
		DeleteContext: resourceObjectStorageEndpointDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(2 * time.Minute),
			Read:   schema.DefaultTimeout(1 * time.Minute),
			Update: schema.DefaultTimeout(2 * time.Minute),
			Delete: schema.DefaultTimeout(2 * time.Minute),
		},
		Schema: schemaObjectStorageEndpoint(),
	}
}

func schemaObjectStorageEndpoint() map[string]*schema.Schema {
	return mergeSchemas(
		schemaObjectStorageCommon(),
		schemaObjectStorageState(),
		map[string]*schema.Schema{
			"url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "URL of the endpoint.",
			},
			"endpoint_user": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the endpoint user.",
			},
			"endpoint_password": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "Password of the endpoint user.",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Indicates if the endpoint is enabled.",
			},
		},
	)
}

func resourceObjectStorageEndpointCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	a := apiFromProviderConfig(m)

	endpoint := objectstoragev2.Endpoint{
		URL:              d.Get("url").(string),
		EndpointUser:     d.Get("endpoint_user").(string),
		EndpointPassword: d.Get("endpoint_password").(string),
		Enabled:          d.Get("enabled").(bool),
	}

	// Set common fields
	setObjectStorageCommonFields(&endpoint, d)
	setObjectStorageStateField(&endpoint, d)

	if err := a.Create(ctx, &endpoint); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(endpoint.Identifier)
	return resourceObjectStorageEndpointRead(ctx, d, m)
}

func resourceObjectStorageEndpointRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	a := apiFromProviderConfig(m)

	endpoint := objectstoragev2.Endpoint{Identifier: d.Id()}
	err := a.Get(ctx, &endpoint)

	if api.IgnoreNotFound(err) != nil {
		return diag.FromErr(err)
	} else if err != nil {
		d.SetId("")
		return nil
	}

	// Set computed and mutable fields
	if err := d.Set("url", endpoint.URL); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("endpoint_user", endpoint.EndpointUser); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("enabled", endpoint.Enabled); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	// Set common fields
	setObjectStorageCommonFieldsFromAPI(&endpoint, d, &diags)
	setObjectStorageStateFieldFromAPI(&endpoint, d, &diags)

	return diags
}

func resourceObjectStorageEndpointUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	a := apiFromProviderConfig(m)

	endpoint := objectstoragev2.Endpoint{
		Identifier:       d.Id(),
		URL:              d.Get("url").(string),
		EndpointUser:     d.Get("endpoint_user").(string),
		EndpointPassword: d.Get("endpoint_password").(string),
		Enabled:          d.Get("enabled").(bool),
	}

	// Set common fields
	setObjectStorageCommonFields(&endpoint, d)
	setObjectStorageStateField(&endpoint, d)

	if err := a.Update(ctx, &endpoint); err != nil {
		return diag.FromErr(err)
	}

	return resourceObjectStorageEndpointRead(ctx, d, m)
}

func resourceObjectStorageEndpointDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	a := apiFromProviderConfig(m)

	endpoint := objectstoragev2.Endpoint{Identifier: d.Id()}
	if err := a.Destroy(ctx, &endpoint); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
