package anxcloud

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"go.anx.io/go-anxcloud/pkg/apis/common"
	objectstoragev2 "go.anx.io/go-anxcloud/pkg/apis/objectstorage/v2"
)

func resourceObjectStorageKey() *schema.Resource {
	return &schema.Resource{
		Description:   "This resource allows you to configure Object Storage keys.",
		CreateContext: resourceObjectStorageKeyCreate,
		ReadContext:   resourceObjectStorageKeyRead,
		UpdateContext: resourceObjectStorageKeyUpdate,
		DeleteContext: resourceObjectStorageKeyDelete,
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
				"name": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The name of the key.",
				},
				"backend": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Identifier of the S3 backend this key belongs to.",
				},
				"tenant": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Identifier of the tenant this key belongs to.",
				},
				"user": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Identifier of the user this key belongs to.",
				},
				"expire_date": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Expiration date for the key in RFC3339 format.",
				},
				"remote_id": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Remote identifier of the key.",
				},
				"secret": {
					Type:        schema.TypeString,
					Computed:    true,
					Sensitive:   true,
					Description: "Secret key for authentication.",
				},
				"secret_url": {
					Type:        schema.TypeString,
					Computed:    true,
					Sensitive:   true,
					Description: "URL containing the secret key.",
				},
			},
		),
	}
}

func resourceObjectStorageKeyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	a := apiFromProviderConfig(m)
	key := objectstoragev2.Key{
		Name: d.Get("name").(string),
	}

	if v, ok := d.GetOk("backend"); ok {
		key.Backend = &common.PartialResource{
			Identifier: v.(string),
		}
	}

	if v, ok := d.GetOk("tenant"); ok {
		key.Tenant = &common.PartialResource{
			Identifier: v.(string),
		}
	}

	if v, ok := d.GetOk("user"); ok {
		key.User = &common.PartialResource{
			Identifier: v.(string),
		}
	}

	if v, ok := d.GetOk("expire_date"); ok {
		expireTime, err := time.Parse(time.RFC3339, v.(string))
		if err != nil {
			return diag.FromErr(err)
		}
		key.ExpireDate = &expireTime
	}

	setObjectStorageCommonFields(&key, d)
	setObjectStorageStateField(&key, d)

	err := a.Create(ctx, &key)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(key.Identifier)
	return resourceObjectStorageKeyRead(ctx, d, m)
}

func resourceObjectStorageKeyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	a := apiFromProviderConfig(m)
	key := objectstoragev2.Key{Identifier: d.Id()}

	err := a.Get(ctx, &key)
	if err != nil {
		if err := handleNotFoundError(err); err != nil {
			return diag.FromErr(err)
		}
		d.SetId("")
		return nil
	}

	if err := d.Set("name", key.Name); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	if key.Backend != nil {
		if err := d.Set("backend", key.Backend.Identifier); err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
	}

	if key.Tenant != nil {
		if err := d.Set("tenant", key.Tenant.Identifier); err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
	}

	if key.User != nil {
		if err := d.Set("user", key.User.Identifier); err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
	}

	if key.ExpireDate != nil {
		if err := d.Set("expire_date", key.ExpireDate.Format(time.RFC3339)); err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
	}

	if key.RemoteID != nil {
		if err := d.Set("remote_id", *key.RemoteID); err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
	}

	if err := d.Set("secret", key.Secret); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	if err := d.Set("secret_url", key.SecretURL); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	setObjectStorageCommonFieldsFromAPI(&key, d, &diags)
	setObjectStorageStateFieldFromAPI(&key, d, &diags)

	return diags
}

func resourceObjectStorageKeyUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	a := apiFromProviderConfig(m)
	key := objectstoragev2.Key{Identifier: d.Id()}

	err := a.Get(ctx, &key)
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("name") {
		key.Name = d.Get("name").(string)
	}

	if d.HasChange("expire_date") {
		if v, ok := d.GetOk("expire_date"); ok {
			expireTime, err := time.Parse(time.RFC3339, v.(string))
			if err != nil {
				return diag.FromErr(err)
			}
			key.ExpireDate = &expireTime
		} else {
			key.ExpireDate = nil
		}
	}

	setObjectStorageCommonFields(&key, d)
	setObjectStorageStateField(&key, d)

	err = a.Update(ctx, &key)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceObjectStorageKeyRead(ctx, d, m)
}

func resourceObjectStorageKeyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	a := apiFromProviderConfig(m)
	key := objectstoragev2.Key{Identifier: d.Id()}

	err := a.Destroy(ctx, &key)
	if err != nil {
		if err := handleNotFoundError(err); err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId("")
	return nil
}
