package anxcloud

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"go.anx.io/go-anxcloud/pkg/apis/common"
	objectstoragev2 "go.anx.io/go-anxcloud/pkg/apis/objectstorage/v2"
)

func resourceObjectStorageRegion() *schema.Resource {
	return &schema.Resource{
		Description:   "This resource allows you to configure Object Storage regions.",
		CreateContext: resourceObjectStorageRegionCreate,
		ReadContext:   resourceObjectStorageRegionRead,
		UpdateContext: resourceObjectStorageRegionUpdate,
		DeleteContext: resourceObjectStorageRegionDelete,
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
					Description: "The name of the region.",
				},
				"description": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "Description of the region.",
				},
				"backend": {
					Type:        schema.TypeString,
					Optional:    true,
					ForceNew:    true,
					Description: "Identifier of the S3 backend this region belongs to.",
				},
			},
		),
	}
}

func resourceObjectStorageRegionCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	a := apiFromProviderConfig(m)
	region := objectstoragev2.Region{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
	}

	if v, ok := d.GetOk("backend"); ok {
		region.Backend = &common.PartialResource{
			Identifier: v.(string),
		}
	}

	setObjectStorageCommonFields(&region, d)
	setObjectStorageStateField(&region, d)

	err := a.Create(ctx, &region)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(region.Identifier)
	return resourceObjectStorageRegionRead(ctx, d, m)
}

func resourceObjectStorageRegionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	a := apiFromProviderConfig(m)
	region := objectstoragev2.Region{Identifier: d.Id()}

	err := a.Get(ctx, &region)
	if err != nil {
		if err := handleNotFoundError(err); err != nil {
			return diag.FromErr(err)
		}
		d.SetId("")
		return nil
	}

	if err := d.Set("name", region.Name); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	if err := d.Set("description", region.Description); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	if region.Backend != nil {
		if err := d.Set("backend", region.Backend.Identifier); err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
	}

	setObjectStorageCommonFieldsFromAPI(&region, d, &diags)
	setObjectStorageStateFieldFromAPI(&region, d, &diags)

	return diags
}

func resourceObjectStorageRegionUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	a := apiFromProviderConfig(m)
	region := objectstoragev2.Region{Identifier: d.Id()}

	err := a.Get(ctx, &region)
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("name") {
		region.Name = d.Get("name").(string)
	}

	if d.HasChange("description") {
		region.Description = d.Get("description").(string)
	}

	setObjectStorageCommonFields(&region, d)
	setObjectStorageStateField(&region, d)

	err = a.Update(ctx, &region)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceObjectStorageRegionRead(ctx, d, m)
}

func resourceObjectStorageRegionDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	a := apiFromProviderConfig(m)
	region := objectstoragev2.Region{Identifier: d.Id()}

	err := a.Destroy(ctx, &region)
	if err != nil {
		if err := handleNotFoundError(err); err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId("")
	return nil
}
