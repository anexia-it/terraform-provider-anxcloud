package anxcloud

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"go.anx.io/go-anxcloud/pkg/api"
	"go.anx.io/go-anxcloud/pkg/apis/common"
	objectstoragev2 "go.anx.io/go-anxcloud/pkg/apis/objectstorage/v2"
)

func resourceObjectStorageBackend() *schema.Resource {
	return &schema.Resource{
		Description:   "This resource allows you to create and manage Object Storage S3 backends.",
		CreateContext: resourceObjectStorageBackendCreate,
		ReadContext:   resourceObjectStorageBackendRead,
		UpdateContext: resourceObjectStorageBackendUpdate,
		DeleteContext: resourceObjectStorageBackendDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(2 * time.Minute),
			Read:   schema.DefaultTimeout(1 * time.Minute),
			Update: schema.DefaultTimeout(2 * time.Minute),
			Delete: schema.DefaultTimeout(2 * time.Minute),
		},
		Schema: schemaObjectStorageBackend(),
	}
}

func schemaObjectStorageBackend() map[string]*schema.Schema {
	return mergeSchemas(
		schemaObjectStorageCommon(),
		schemaObjectStorageState(),
		map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the S3 backend.",
			},
			"endpoint": schemaObjectStorageReference(),
			"backend_type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "1",
				ValidateFunc: validation.StringInSlice([]string{
					"1", // NetApp Storage Grid v4
				}, false),
				Description: "Type of the S3 backend. 1 = NetApp Storage Grid v4.",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Indicates if the S3 backend is enabled.",
			},
			"backend_user": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the S3 backend user.",
			},
			"backend_password": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "Password of the S3 backend user.",
			},
		},
	)
}

func resourceObjectStorageBackendCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	a := apiFromProviderConfig(m)

	backendType := &objectstoragev2.GenericAttributeSelect{
		Identifier: d.Get("backend_type").(string),
	}

	enabled := d.Get("enabled").(bool)
	backend := objectstoragev2.S3Backend{
		Name: d.Get("name").(string),
		Endpoint: common.PartialResource{
			Identifier: d.Get("endpoint").(string),
		},
		BackendType:     backendType,
		Enabled:         &enabled,
		BackendUser:     d.Get("backend_user").(string),
		BackendPassword: d.Get("backend_password").(string),
	}

	// Set common fields
	setObjectStorageCommonFields(&backend, d)
	setObjectStorageStateField(&backend, d)

	if err := a.Create(ctx, &backend); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(backend.Identifier)
	return resourceObjectStorageBackendRead(ctx, d, m)
}

func resourceObjectStorageBackendRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	a := apiFromProviderConfig(m)

	backend := objectstoragev2.S3Backend{Identifier: d.Id()}
	err := a.Get(ctx, &backend)

	if api.IgnoreNotFound(err) != nil {
		return diag.FromErr(err)
	} else if err != nil {
		d.SetId("")
		return nil
	}

	// Set computed and mutable fields
	if err := d.Set("name", backend.Name); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("endpoint", backend.Endpoint.Identifier); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if backend.BackendType != nil {
		if err := d.Set("backend_type", backend.BackendType.Identifier); err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
	}
	if backend.Enabled != nil {
		if err := d.Set("enabled", *backend.Enabled); err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
	}
	if err := d.Set("backend_user", backend.BackendUser); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	// Set common fields
	setObjectStorageCommonFieldsFromAPI(&backend, d, &diags)
	setObjectStorageStateFieldFromAPI(&backend, d, &diags)

	return diags
}

func resourceObjectStorageBackendUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	a := apiFromProviderConfig(m)

	backendType := &objectstoragev2.GenericAttributeSelect{
		Identifier: d.Get("backend_type").(string),
	}

	enabled := d.Get("enabled").(bool)
	backend := objectstoragev2.S3Backend{
		Identifier: d.Id(),
		Name:       d.Get("name").(string),
		Endpoint: common.PartialResource{
			Identifier: d.Get("endpoint").(string),
		},
		BackendType:     backendType,
		Enabled:         &enabled,
		BackendUser:     d.Get("backend_user").(string),
		BackendPassword: d.Get("backend_password").(string),
	}

	// Set common fields
	setObjectStorageCommonFields(&backend, d)
	setObjectStorageStateField(&backend, d)

	if err := a.Update(ctx, &backend); err != nil {
		return diag.FromErr(err)
	}

	return resourceObjectStorageBackendRead(ctx, d, m)
}

func resourceObjectStorageBackendDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	a := apiFromProviderConfig(m)

	backend := objectstoragev2.S3Backend{Identifier: d.Id()}
	if err := a.Destroy(ctx, &backend); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
