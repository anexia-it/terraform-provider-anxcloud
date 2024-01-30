package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.anx.io/go-anxcloud/pkg/api"
	corev1 "go.anx.io/go-anxcloud/pkg/apis/core/v1"
)

func ensureTags(ctx context.Context, engine api.API, id string, plan tfsdk.Plan) (diags diag.Diagnostics) {
	var tagSet types.Set
	diags.Append(plan.GetAttribute(ctx, path.Root("tags"), &tagSet)...)
	if diags.HasError() {
		return
	}

	var tags []string
	diags.Append(tagSet.ElementsAs(ctx, &tags, true)...)

	resource := corev1.Resource{Identifier: id}

	remote, err := corev1.ListTags(ctx, engine, &resource)
	if err != nil {
		diags.AddError("Unable to list tags", err.Error())
		return
	}

	toRemove := sliceSubstract(remote, tags)
	if err := corev1.Untag(ctx, engine, &resource, toRemove...); err != nil {
		diags.AddError("Failed to untag resource", err.Error())
	}

	toAdd := sliceSubstract(tags, remote)
	if err := corev1.Tag(ctx, engine, &resource, toAdd...); err != nil {
		diags.AddError("Failed to tag resource", err.Error())
	}

	return
}

func sliceSubstract[T comparable](a, b []T) []T {
	out := make([]T, 0, len(a))
outer:
	for i := range a {
		for j := range b {
			if a[i] == b[j] {
				continue outer
			}
		}
		out = append(out, a[i])
	}
	return out
}

func readTags(ctx context.Context, engine api.API, id string, tagSet *types.Set) (diags diag.Diagnostics) {
	tags, err := corev1.ListTags(ctx, engine, &corev1.Resource{Identifier: id})
	if err != nil {
		diags.AddError("Unable to list tags", err.Error())
		return
	}

	newTagSet, tagSetDiags := types.SetValueFrom(ctx, types.StringType, &tags)
	diags.Append(tagSetDiags...)

	*tagSet = newTagSet

	return
}
