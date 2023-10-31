package anxcloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"go.anx.io/go-anxcloud/pkg/api"
	frontierv1 "go.anx.io/go-anxcloud/pkg/apis/frontier/v1"
	"slices"
)

func frontierActionTypesExcept(exception frontierv1.ActionType) []string {
	types := []string{
		string(frontierv1.ActionTypeURLRewrite),
		string(frontierv1.ActionTypeMockResponse),
		string(frontierv1.ActionTypeE5EFunction),
		string(frontierv1.ActionTypeE5EAsyncFunction),
		string(frontierv1.ActionTypeE5EAsyncResult),
	}

	return slices.DeleteFunc(types, func(actionType string) bool {
		return actionType == string(exception)
	})
}

func resourceFrontierAction() *schema.Resource {
	return &schema.Resource{
		Description: "An action is the lowest entity within Frontier's hierarchy and maps HTTP methods for an endpoint to action handlers." +
			" Those action handlers may be e5e functions, other HTTP-based APIs or mock responses." +
			" Referencing a non-existing e5e function will result in a 404 error.",
		CreateContext: resourceFrontierActionCreate,
		ReadContext:   resourceFrontierActionRead,
		UpdateContext: resourceFrontierActionUpdate,
		DeleteContext: resourceFrontierActionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Action identifier.",
			},
			"endpoint": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Action endpoint identifier.",
			},
			"http_request_method": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Action HTTP request method.",
			},
			"url_rewrite": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"url": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
				ConflictsWith: frontierActionTypesExcept(frontierv1.ActionTypeURLRewrite),
			},
			"mock_response": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"body": {
							Type:     schema.TypeString,
							Required: true,
						},
						"language": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
				ConflictsWith: frontierActionTypesExcept(frontierv1.ActionTypeMockResponse),
			},
			"e5e_function": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"function": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
				ConflictsWith: frontierActionTypesExcept(frontierv1.ActionTypeE5EFunction),
			},
			"e5e_async_function": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"function": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
				ConflictsWith: frontierActionTypesExcept(frontierv1.ActionTypeE5EAsyncFunction),
			},
			"e5e_async_result": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"function": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
				ConflictsWith: frontierActionTypesExcept(frontierv1.ActionTypeE5EAsyncResult),
			},
		},
	}
}

func frontierActionFromResourceData(d *schema.ResourceData) frontierv1.Action {
	frontierAction := frontierv1.Action{
		Identifier:         d.Id(),
		EndpointIdentifier: d.Get("endpoint").(string),
		HTTPRequestMethod:  d.Get("http_request_method").(string),
		Meta:               &frontierv1.ActionMeta{},
	}

	getMeta := func(key string) (map[string]any, bool) {
		if meta, ok := d.GetOk(key); ok {
			return meta.([]any)[0].(map[string]any), true
		}
		return nil, false
	}

	if meta, ok := getMeta("mock_response"); ok {
		frontierAction.Type = frontierv1.ActionTypeMockResponse
		frontierAction.Meta.ActionMetaMockResponse = &frontierv1.ActionMetaMockResponse{
			Body:     meta["body"].(string),
			Language: meta["language"].(string),
		}
	} else if meta, ok := getMeta("url_rewrite"); ok {
		frontierAction.Type = frontierv1.ActionTypeURLRewrite
		frontierAction.Meta.ActionMetaURLRewrite = &frontierv1.ActionMetaURLRewrite{
			URL: meta["url"].(string),
		}
	} else if meta, ok := getMeta("e5e_function"); ok {
		frontierAction.Type = frontierv1.ActionTypeE5EFunction
		frontierAction.Meta.ActionMetaE5EFunction = &frontierv1.ActionMetaE5EFunction{
			FunctionIdentifier: meta["function"].(string),
		}
	} else if meta, ok := getMeta("e5e_async_function"); ok {
		frontierAction.Type = frontierv1.ActionTypeE5EAsyncFunction
		frontierAction.Meta.ActionMetaE5EAsyncFunction = &frontierv1.ActionMetaE5EAsyncFunction{
			FunctionIdentifier: meta["function"].(string),
		}
	} else if meta, ok := getMeta("e5e_async_result"); ok {
		frontierAction.Type = frontierv1.ActionTypeE5EAsyncResult
		frontierAction.Meta.ActionMetaE5EAsyncResult = &frontierv1.ActionMetaE5EAsyncResult{
			FunctionIdentifier: meta["function"].(string),
		}
	}

	return frontierAction
}
func resourceFrontierActionCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	a := apiFromProviderConfig(m)

	frontierAction := frontierActionFromResourceData(d)
	if err := a.Create(ctx, &frontierAction); err != nil {
		return diag.Errorf("failed to create resource: %s", err)
	}

	d.SetId(frontierAction.Identifier)

	return resourceFrontierActionRead(ctx, d, m)
}

func resourceFrontierActionRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	a := apiFromProviderConfig(m)

	frontierAction := frontierv1.Action{Identifier: d.Id()}
	if err := a.Get(ctx, &frontierAction); api.IgnoreNotFound(err) != nil {
		return diag.Errorf("failed getting resource: %s", err)
	} else if err != nil {
		d.SetId("")
		return nil
	}

	var diags diag.Diagnostics

	if err := d.Set("endpoint", frontierAction.EndpointIdentifier); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("http_request_method", frontierAction.HTTPRequestMethod); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	var metaValue []any
	switch frontierAction.Type {
	case frontierv1.ActionTypeMockResponse:
		metaValue = []any{map[string]any{
			"body":     frontierAction.Meta.ActionMetaMockResponse.Body,
			"language": frontierAction.Meta.ActionMetaMockResponse.Language,
		}}
	case frontierv1.ActionTypeURLRewrite:
		metaValue = []any{map[string]any{
			"url": frontierAction.Meta.ActionMetaURLRewrite.URL,
		}}
	case frontierv1.ActionTypeE5EFunction:
		metaValue = []any{map[string]any{
			"function": frontierAction.Meta.ActionMetaE5EFunction.FunctionIdentifier,
		}}
	case frontierv1.ActionTypeE5EAsyncFunction:
		metaValue = []any{map[string]any{
			"function": frontierAction.Meta.ActionMetaE5EAsyncFunction.FunctionIdentifier,
		}}
	case frontierv1.ActionTypeE5EAsyncResult:
		metaValue = []any{map[string]any{
			"function": frontierAction.Meta.ActionMetaE5EAsyncResult.FunctionIdentifier,
		}}
	}

	if err := d.Set(string(frontierAction.Type), metaValue); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	return diags
}

func resourceFrontierActionUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	a := apiFromProviderConfig(m)

	frontierAction := frontierActionFromResourceData(d)
	if err := a.Update(ctx, &frontierAction); err != nil {
		return diag.Errorf("failed updating resource: %s", err)
	}

	return resourceFrontierActionRead(ctx, d, m)
}

func resourceFrontierActionDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	a := apiFromProviderConfig(m)

	if err := a.Destroy(ctx, &frontierv1.Action{Identifier: d.Id()}); api.IgnoreNotFound(err) != nil {
		return diag.Errorf("failed deleting resource: %s", err)
	}

	return nil
}
