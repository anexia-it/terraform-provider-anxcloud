package anxcloud

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"go.anx.io/go-anxcloud/pkg/api"
	"go.anx.io/go-anxcloud/pkg/apis/common"
	objectstoragev2 "go.anx.io/go-anxcloud/pkg/apis/objectstorage/v2"
)

func resourceObjectStorageBucket() *schema.Resource {
	return &schema.Resource{
		Description:   "This resource allows you to create and manage Object Storage buckets.",
		CreateContext: resourceObjectStorageBucketCreate,
		ReadContext:   resourceObjectStorageBucketRead,
		UpdateContext: resourceObjectStorageBucketUpdate,
		DeleteContext: resourceObjectStorageBucketDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(2 * time.Minute),
			Read:   schema.DefaultTimeout(1 * time.Minute),
			Update: schema.DefaultTimeout(2 * time.Minute),
			Delete: schema.DefaultTimeout(2 * time.Minute),
		},
		Schema: schemaObjectStorageBucket(),
	}
}

func schemaObjectStorageBucket() map[string]*schema.Schema {
	return mergeSchemas(
		schemaObjectStorageCommon(),
		schemaObjectStorageState(),
		map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the bucket.",
			},
			"region":  schemaObjectStorageReference(),
			"backend": schemaObjectStorageReference(),
			"tenant":  schemaObjectStorageReference(),
			"object_count": {
				Type:        schema.TypeFloat,
				Computed:    true,
				Description: "Number of objects in the bucket.",
			},
			"object_size": {
				Type:        schema.TypeFloat,
				Computed:    true,
				Description: "Collective size of objects in the bucket.",
			},
			"object_lock_lifetime_in_days": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Number of days for object lock lifetime. When set, objects in this bucket will be protected from deletion and modification for the specified duration.",
			},
		},
	)
}

func resourceObjectStorageBucketCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	a := apiFromProviderConfig(m)

	bucket := objectstoragev2.Bucket{
		Name: d.Get("name").(string),
		Region: common.PartialResource{
			Identifier: d.Get("region").(string),
		},
		Backend: common.PartialResource{
			Identifier: d.Get("backend").(string),
		},
		Tenant: common.PartialResource{
			Identifier: d.Get("tenant").(string),
		},
	}

	if v, ok := d.GetOk("object_lock_lifetime_in_days"); ok {
		lifetime := v.(int)
		bucket.ObjectLockLifetime = &lifetime
	}

	// Set common fields
	setObjectStorageCommonFields(&bucket, d)
	setObjectStorageStateField(&bucket, d)

	if err := a.Create(ctx, &bucket); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(bucket.Identifier)
	return resourceObjectStorageBucketRead(ctx, d, m)
}

func resourceObjectStorageBucketRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	a := apiFromProviderConfig(m)

	bucket := objectstoragev2.Bucket{Identifier: d.Id()}
	err := a.Get(ctx, &bucket)

	if api.IgnoreNotFound(err) != nil {
		return diag.FromErr(err)
	} else if err != nil {
		d.SetId("")
		return nil
	}

	// Set computed and mutable fields
	if err := d.Set("name", bucket.Name); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("region", bucket.Region.Identifier); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("backend", bucket.Backend.Identifier); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("tenant", bucket.Tenant.Identifier); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	if bucket.ObjectLockLifetime != nil {
		if err := d.Set("object_lock_lifetime_in_days", *bucket.ObjectLockLifetime); err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
	}

	if objectCount, err := bucket.GetObjectCount(); err == nil {
		if err := d.Set("object_count", objectCount); err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
	}
	if objectSize, err := bucket.GetObjectSize(); err == nil {
		if err := d.Set("object_size", objectSize); err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
	}

	// Set common fields
	setObjectStorageCommonFieldsFromAPI(&bucket, d, &diags)
	setObjectStorageStateFieldFromAPI(&bucket, d, &diags)

	return diags
}

func resourceObjectStorageBucketUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	a := apiFromProviderConfig(m)

	bucket := objectstoragev2.Bucket{
		Identifier: d.Id(),
		Name:       d.Get("name").(string),
		Region: common.PartialResource{
			Identifier: d.Get("region").(string),
		},
		Backend: common.PartialResource{
			Identifier: d.Get("backend").(string),
		},
		Tenant: common.PartialResource{
			Identifier: d.Get("tenant").(string),
		},
	}

	if v, ok := d.GetOk("object_lock_lifetime_in_days"); ok {
		lifetime := v.(int)
		bucket.ObjectLockLifetime = &lifetime
	}

	// Set common fields
	setObjectStorageCommonFields(&bucket, d)
	setObjectStorageStateField(&bucket, d)

	if err := a.Update(ctx, &bucket); err != nil {
		return diag.FromErr(err)
	}

	return resourceObjectStorageBucketRead(ctx, d, m)
}

func resourceObjectStorageBucketDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	a := apiFromProviderConfig(m)

	bucket := objectstoragev2.Bucket{Identifier: d.Id()}
	if err := a.Destroy(ctx, &bucket); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
