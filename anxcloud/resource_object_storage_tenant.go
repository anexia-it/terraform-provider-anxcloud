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

func resourceObjectStorageTenant() *schema.Resource {
	return &schema.Resource{
		Description:   "This resource allows you to create and manage Object Storage tenants.",
		CreateContext: resourceObjectStorageTenantCreate,
		ReadContext:   resourceObjectStorageTenantRead,
		UpdateContext: resourceObjectStorageTenantUpdate,
		DeleteContext: resourceObjectStorageTenantDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(2 * time.Minute),
			Read:   schema.DefaultTimeout(1 * time.Minute),
			Update: schema.DefaultTimeout(2 * time.Minute),
			Delete: schema.DefaultTimeout(2 * time.Minute),
		},
		Schema: schemaObjectStorageTenant(),
	}
}

func schemaObjectStorageTenant() map[string]*schema.Schema {
	return mergeSchemas(
		schemaObjectStorageCommon(),
		schemaObjectStorageState(),
		map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the tenant.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the tenant (can be empty).",
			},
			"user_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the tenant user to be used for API login.",
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "Password of the tenant user to be used for API login.",
			},
			"quota": {
				Type:         schema.TypeFloat,
				Required:     true,
				ValidateFunc: validation.FloatAtLeast(1),
				Description:  "Maximum number of bytes allowed for objects within buckets (must be > 0).",
			},
			"usage": {
				Type:        schema.TypeFloat,
				Computed:    true,
				Description: "Current number of bytes used by the user within buckets.",
			},
			"backend": schemaObjectStorageReference(),
			"remote_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Tenant ID in the backend system (mandatory).",
			},
		},
	)
}

func resourceObjectStorageTenantCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	a := apiFromProviderConfig(m)

	tenant := objectstoragev2.Tenant{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		UserName:    d.Get("user_name").(string),
		Password:    d.Get("password").(string),
		Backend: common.PartialResource{
			Identifier: d.Get("backend").(string),
		},
	}

	quota := d.Get("quota").(float64)
	tenant.Quota = &quota

	remoteID := d.Get("remote_id").(string)
	tenant.RemoteID = &remoteID

	// Set common fields
	setObjectStorageCommonFields(&tenant, d)
	setObjectStorageStateField(&tenant, d)

	if err := a.Create(ctx, &tenant); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(tenant.Identifier)
	return resourceObjectStorageTenantRead(ctx, d, m)
}

func resourceObjectStorageTenantRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	a := apiFromProviderConfig(m)

	tenant := objectstoragev2.Tenant{Identifier: d.Id()}
	err := a.Get(ctx, &tenant)

	if api.IgnoreNotFound(err) != nil {
		return diag.FromErr(err)
	} else if err != nil {
		d.SetId("")
		return nil
	}

	// Set computed and mutable fields
	if err := d.Set("name", tenant.Name); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("description", tenant.Description); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("user_name", tenant.UserName); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("backend", tenant.Backend.Identifier); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	if tenant.Quota != nil {
		if err := d.Set("quota", *tenant.Quota); err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
	}
	if tenant.Usage != nil {
		if err := d.Set("usage", *tenant.Usage); err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
	}
	if tenant.RemoteID != nil {
		if err := d.Set("remote_id", *tenant.RemoteID); err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
	} else {
		// This shouldn't happen since remote_id is required, but handle gracefully
		if err := d.Set("remote_id", ""); err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
	}

	// Set common fields
	setObjectStorageCommonFieldsFromAPI(&tenant, d, &diags)
	setObjectStorageStateFieldFromAPI(&tenant, d, &diags)

	return diags
}

func resourceObjectStorageTenantUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	a := apiFromProviderConfig(m)

	tenant := objectstoragev2.Tenant{
		Identifier:  d.Id(),
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		UserName:    d.Get("user_name").(string),
		Password:    d.Get("password").(string),
		Backend: common.PartialResource{
			Identifier: d.Get("backend").(string),
		},
	}

	quota := d.Get("quota").(float64)
	tenant.Quota = &quota

	remoteID := d.Get("remote_id").(string)
	tenant.RemoteID = &remoteID

	// Set common fields
	setObjectStorageCommonFields(&tenant, d)
	setObjectStorageStateField(&tenant, d)

	if err := a.Update(ctx, &tenant); err != nil {
		return diag.FromErr(err)
	}

	return resourceObjectStorageTenantRead(ctx, d, m)
}

func resourceObjectStorageTenantDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	a := apiFromProviderConfig(m)

	tenant := objectstoragev2.Tenant{Identifier: d.Id()}
	if err := a.Destroy(ctx, &tenant); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
