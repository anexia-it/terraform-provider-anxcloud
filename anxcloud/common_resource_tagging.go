package anxcloud

import (
	"context"
	"fmt"
	"reflect"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"go.anx.io/go-anxcloud/pkg/api"
	corev1 "go.anx.io/go-anxcloud/pkg/apis/core/v1"
)

func withTagsAttribute(s schemaMap) schemaMap {
	s["tags"] = &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		Computed:    true,
		Description: "List of tags attached to the resource.",
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
		// suppress diff when only the order has changed
		DiffSuppressOnRefresh: true,
		DiffSuppressFunc: func(k, oldValue, newValue string, d *schema.ResourceData) bool {
			o, n := d.GetChange("tags")
			oStringArray := mustCastInterfaceArray[string](o.([]interface{}))
			nStringArray := mustCastInterfaceArray[string](n.([]interface{}))
			sort.Strings(oStringArray)
			sort.Strings(nStringArray)
			return reflect.DeepEqual(oStringArray, nStringArray)
		},
	}
	return s
}

type schemaContextCreateOrUpdateFunc interface {
	schema.CreateContextFunc | schema.UpdateContextFunc
}

func tagsMiddlewareCreate(wrapped schema.CreateContextFunc) schema.CreateContextFunc {
	return ensureTagsMiddleware(wrapped)
}

func tagsMiddlewareUpdate(wrapped schema.UpdateContextFunc) schema.UpdateContextFunc {
	return ensureTagsMiddleware(wrapped)
}

func ensureTagsMiddleware[T schemaContextCreateOrUpdateFunc](wrapped T) T {
	return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		diags := wrapped(ctx, d, m)
		if diags.HasError() {
			return diags
		}

		// we don't touch remote tags when tags attribute is not set
		// remote tags are also kept when tags attribute was unset
		if d.GetRawConfig().GetAttr("tags").IsNull() {
			return diags
		}

		tags := mustCastInterfaceArray[string](d.Get("tags").([]interface{}))
		if err := ensureTags(ctx, m.(providerContext).api, d.Id(), tags); err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}

		return diags
	}
}

func tagsMiddlewareRead(wrapped schema.ReadContextFunc) schema.ReadContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		diags := wrapped(ctx, d, m)
		// Resources id can be zero-val after read if remote resource
		// was deleted manually via engine, but is still present in tf state.
		// Because reading tags from a non-existing resource fails,
		// we want to skip tagging logic.
		if diags.HasError() || d.Id() == "" {
			return diags
		}

		tags, err := readTags(ctx, m.(providerContext).api, d.Id())
		if err != nil {
			return append(diags, diag.FromErr(err)...)
		}

		if err := d.Set("tags", tags); err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}

		return diags
	}
}

func ensureTags(ctx context.Context, a api.API, resourceID string, tags []string) error {
	resource := corev1.Resource{Identifier: resourceID}

	remote, err := readTags(ctx, a, resourceID)
	if err != nil {
		return fmt.Errorf("failed to fetch remote tags: %w", err)
	}

	toRemove := sliceSubstract(remote, tags)
	if err := corev1.Untag(ctx, a, &resource, toRemove...); err != nil {
		return fmt.Errorf("failed to untag resource: %w", err)
	}

	toAdd := sliceSubstract(tags, remote)
	if err := corev1.Tag(ctx, a, &resource, toAdd...); err != nil {
		return fmt.Errorf("failed to tag resource: %w", err)
	}

	return nil
}

func readTags(ctx context.Context, a api.API, resourceID string) ([]string, error) {
	return corev1.ListTags(ctx, a, &corev1.Resource{Identifier: resourceID})
}
