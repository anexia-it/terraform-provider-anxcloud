package planmodifiers

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

type keepStringPrefixModifier struct{}

func (m keepStringPrefixModifier) Description(_ context.Context) string {
	return "Ensures that if the the plan value is the suffix of the state value, the value from state will be preserved"
}

func (m keepStringPrefixModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m keepStringPrefixModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}

	if strings.HasSuffix(req.StateValue.ValueString(), req.PlanValue.ValueString()) {
		resp.PlanValue = req.StateValue
	}
}

func KeepStringPrefix() planmodifier.String {
	return keepStringPrefixModifier{}
}

type keepStringSuffixModifier struct{}

func (m keepStringSuffixModifier) Description(_ context.Context) string {
	return "Ensures that if the the plan value is the prefix of the state value, the value from state will be preserved"
}

func (m keepStringSuffixModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m keepStringSuffixModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}

	if strings.HasPrefix(req.StateValue.ValueString(), req.PlanValue.ValueString()) {
		resp.PlanValue = req.StateValue
	}
}

func KeepStringSuffix() planmodifier.String {
	return keepStringSuffixModifier{}
}
