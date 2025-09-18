package anxcloud

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"go.anx.io/go-anxcloud/pkg/apis/common"
	objectstoragev2 "go.anx.io/go-anxcloud/pkg/apis/objectstorage/v2"
)

func resourceObjectStorageUser() *schema.Resource {
	return &schema.Resource{
		Description:   "This resource allows you to configure Object Storage users.",
		CreateContext: resourceObjectStorageUserCreate,
		ReadContext:   resourceObjectStorageUserRead,
		UpdateContext: resourceObjectStorageUserUpdate,
		DeleteContext: resourceObjectStorageUserDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Read:   schema.DefaultTimeout(1 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},
		Schema: mergeSchemas(
			schemaObjectStorageCommon(),
			schemaObjectStorageState(),
			map[string]*schema.Schema{
				"user_name": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The name of the user.",
				},
				"full_name": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The full name of the user.",
				},
				"enabled": {
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     true,
					Description: "Indicates whether the user is enabled.",
				},
				"backend": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "Identifier of the S3 backend this user belongs to.",
				},
				"tenant": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "Identifier of the tenant this user belongs to.",
				},
				"remote_id": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Remote identifier of the user.",
				},
			},
		),
	}
}

func resourceObjectStorageUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	a := apiFromProviderConfig(m)
	user := objectstoragev2.User{
		UserName: d.Get("user_name").(string),
		FullName: d.Get("full_name").(string),
		Backend: common.PartialResource{
			Identifier: d.Get("backend").(string),
		},
		Tenant: common.PartialResource{
			Identifier: d.Get("tenant").(string),
		},
	}

	if v, ok := d.GetOk("enabled"); ok {
		enabled := v.(bool)
		user.Enabled = &enabled
	}

	setObjectStorageCommonFields(&user, d)
	setObjectStorageStateField(&user, d)

	err := a.Create(ctx, &user)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(user.Identifier)
	return resourceObjectStorageUserRead(ctx, d, m)
}

func resourceObjectStorageUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	a := apiFromProviderConfig(m)
	user := objectstoragev2.User{Identifier: d.Id()}

	err := a.Get(ctx, &user)
	if err != nil {
		if err := handleNotFoundError(err); err != nil {
			return diag.FromErr(err)
		}
		d.SetId("")
		return nil
	}

	if err := d.Set("user_name", user.UserName); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	if err := d.Set("full_name", user.FullName); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	if user.Enabled != nil {
		if err := d.Set("enabled", *user.Enabled); err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
	}

	// Skip setting backend as this is a "write once" field
	// and the API's get endpoint doesn't return it properly

	if err := d.Set("tenant", user.Tenant.Identifier); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	if user.RemoteID != nil {
		if err := d.Set("remote_id", *user.RemoteID); err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
	}

	setObjectStorageCommonFieldsFromAPI(&user, d, &diags)
	setObjectStorageStateFieldFromAPI(&user, d, &diags)

	return diags
}

func resourceObjectStorageUserUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	a := apiFromProviderConfig(m)
	user := objectstoragev2.User{Identifier: d.Id()}

	err := a.Get(ctx, &user)
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("user_name") {
		user.UserName = d.Get("user_name").(string)
	}

	if d.HasChange("full_name") {
		user.FullName = d.Get("full_name").(string)
	}

	if d.HasChange("enabled") {
		enabled := d.Get("enabled").(bool)
		user.Enabled = &enabled
	}

	setObjectStorageCommonFields(&user, d)
	setObjectStorageStateField(&user, d)

	err = a.Update(ctx, &user)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceObjectStorageUserRead(ctx, d, m)
}

func resourceObjectStorageUserDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	a := apiFromProviderConfig(m)
	user := objectstoragev2.User{Identifier: d.Id()}

	err := a.Destroy(ctx, &user)
	if err != nil {
		if err := handleNotFoundError(err); err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId("")
	return nil
}
