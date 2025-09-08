package anxcloud

import (
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	objectstoragev2 "go.anx.io/go-anxcloud/pkg/apis/objectstorage/v2"
)

// schemaObjectStorageCommon returns the common schema fields for all Object Storage resources
func schemaObjectStorageCommon() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"customer": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Identifier of the customer the resource should be assigned to (mandatory).",
		},
		"reseller": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Identifier of the reseller the resource should be assigned to.",
		},
		"share": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Defines whether a resource is shared.",
		},
		"resource_pools": {
			Type:        schema.TypeList,
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Description: "List of resource pool identifiers to which the resource should be assigned.",
		},
		"created_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The creation time of the resource.",
		},
		"updated_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The last update time of the resource.",
		},
	}
}

// schemaObjectStorageState returns the schema for state attribute
func schemaObjectStorageState() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"state": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
			ValidateFunc: validation.StringInSlice([]string{
				"0", "1",
			}, false),
			Description: "State of the resource. 0 = OK, 1 = Error.",
		},
	}
}

// schemaObjectStorageReference returns schema for a reference to another resource
func schemaObjectStorageReference() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: "Identifier of the referenced resource.",
	}
}

// mergeSchemas combines multiple schema maps
func mergeSchemas(schemas ...map[string]*schema.Schema) map[string]*schema.Schema {
	result := make(map[string]*schema.Schema)
	for _, s := range schemas {
		for k, v := range s {
			result[k] = v
		}
	}
	return result
}

// setObjectStorageCommonFields sets the common fields from resource data to any Object Storage struct
func setObjectStorageCommonFields(obj interface{}, d *schema.ResourceData) {
	// Use reflection-like approach by casting to specific interface
	switch o := obj.(type) {
	case *objectstoragev2.Endpoint:
		o.CustomerIdentifier = d.Get("customer").(string)
		if v, ok := d.GetOk("reseller"); ok {
			o.ResellerIdentifier = v.(string)
		}
		if v, ok := d.GetOk("share"); ok {
			o.Share = v.(bool)
		}
	case *objectstoragev2.S3Backend:
		o.CustomerIdentifier = d.Get("customer").(string)
		if v, ok := d.GetOk("reseller"); ok {
			o.ResellerIdentifier = v.(string)
		}
		if v, ok := d.GetOk("share"); ok {
			o.Share = v.(bool)
		}
	case *objectstoragev2.Tenant:
		o.CustomerIdentifier = d.Get("customer").(string)
		if v, ok := d.GetOk("reseller"); ok {
			o.ResellerIdentifier = v.(string)
		}
		if v, ok := d.GetOk("share"); ok {
			o.Share = v.(bool)
		}
	case *objectstoragev2.Bucket:
		o.CustomerIdentifier = d.Get("customer").(string)
		if v, ok := d.GetOk("reseller"); ok {
			o.ResellerIdentifier = v.(string)
		}
		if v, ok := d.GetOk("share"); ok {
			o.Share = v.(bool)
		}
	case *objectstoragev2.User:
		o.CustomerIdentifier = d.Get("customer").(string)
		if v, ok := d.GetOk("reseller"); ok {
			o.ResellerIdentifier = v.(string)
		}
		if v, ok := d.GetOk("share"); ok {
			o.Share = v.(bool)
		}
	case *objectstoragev2.Key:
		o.CustomerIdentifier = d.Get("customer").(string)
		if v, ok := d.GetOk("reseller"); ok {
			o.ResellerIdentifier = v.(string)
		}
		if v, ok := d.GetOk("share"); ok {
			o.Share = v.(bool)
		}
	case *objectstoragev2.Region:
		o.CustomerIdentifier = d.Get("customer").(string)
		if v, ok := d.GetOk("reseller"); ok {
			o.ResellerIdentifier = v.(string)
		}
		if v, ok := d.GetOk("share"); ok {
			o.Share = v.(bool)
		}
	}
}

// setObjectStorageStateField sets the state field from resource data
func setObjectStorageStateField(obj interface{}, d *schema.ResourceData) {
	if v, ok := d.GetOk("state"); ok {
		stateValue := v.(string)
		if stateValue != "" {
			state := &objectstoragev2.GenericAttributeState{
				ID: stateValue,
			}
			switch o := obj.(type) {
			case *objectstoragev2.Endpoint:
				o.State = state
			case *objectstoragev2.S3Backend:
				o.State = state
			case *objectstoragev2.Tenant:
				o.State = state
			case *objectstoragev2.Bucket:
				o.State = state
			case *objectstoragev2.User:
				o.State = state
			case *objectstoragev2.Key:
				o.State = state
			case *objectstoragev2.Region:
				o.State = state
			}
		}
	}
}

// setObjectStorageCommonFieldsFromAPI sets the common fields from API response to resource data
func setObjectStorageCommonFieldsFromAPI(obj interface{}, d *schema.ResourceData, diags *diag.Diagnostics) {
	switch o := obj.(type) {
	case *objectstoragev2.Endpoint:
		if err := d.Set("share", o.Share); err != nil {
			*diags = append(*diags, diag.FromErr(err)...)
		}
		// Note: We don't update customer/reseller from API response to preserve user input
		// The API may return different values (names vs identifiers), but we keep the user's configuration
	case *objectstoragev2.S3Backend:
		if err := d.Set("share", o.Share); err != nil {
			*diags = append(*diags, diag.FromErr(err)...)
		}
		// Note: We don't update customer/reseller from API response to preserve user input
	case *objectstoragev2.Tenant:
		if err := d.Set("share", o.Share); err != nil {
			*diags = append(*diags, diag.FromErr(err)...)
		}
		// Note: We don't update customer/reseller from API response to preserve user input
	case *objectstoragev2.Bucket:
		if err := d.Set("share", o.Share); err != nil {
			*diags = append(*diags, diag.FromErr(err)...)
		}
		// Note: We don't update customer/reseller from API response to preserve user input
	case *objectstoragev2.User:
		if err := d.Set("share", o.Share); err != nil {
			*diags = append(*diags, diag.FromErr(err)...)
		}
		// Note: We don't update customer/reseller from API response to preserve user input
	case *objectstoragev2.Key:
		if err := d.Set("share", o.Share); err != nil {
			*diags = append(*diags, diag.FromErr(err)...)
		}
		// Note: We don't update customer/reseller from API response to preserve user input
	case *objectstoragev2.Region:
		if err := d.Set("share", o.Share); err != nil {
			*diags = append(*diags, diag.FromErr(err)...)
		}
		// Note: We don't update customer/reseller from API response to preserve user input
	}
}

// setObjectStorageStateFieldFromAPI sets the state field from API response to resource data
func setObjectStorageStateFieldFromAPI(obj interface{}, d *schema.ResourceData, diags *diag.Diagnostics) {
	var state *objectstoragev2.GenericAttributeState
	switch o := obj.(type) {
	case *objectstoragev2.Endpoint:
		state = o.State
	case *objectstoragev2.S3Backend:
		state = o.State
	case *objectstoragev2.Tenant:
		state = o.State
	case *objectstoragev2.Bucket:
		state = o.State
	case *objectstoragev2.User:
		state = o.State
	case *objectstoragev2.Key:
		state = o.State
	case *objectstoragev2.Region:
		state = o.State
	}

	if state != nil {
		if err := d.Set("state", state.ID); err != nil {
			*diags = append(*diags, diag.FromErr(err)...)
		}
	}
}

// generateDataSourceID creates a simple ID for data sources based on time
func generateDataSourceID() string {
	return strconv.FormatInt(time.Now().Unix(), 10)
}
