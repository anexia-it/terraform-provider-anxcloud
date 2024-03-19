package planmodifiers

import (
	"context"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

func KeepIPAddressOrderPlanModifier() planmodifier.List {
	return &keepIPAddressOrderPlanModifier{}
}

type keepIPAddressOrderPlanModifier struct{}

func (*keepIPAddressOrderPlanModifier) Description(context.Context) string {
	return "Ensures that if the addresses in state are equal to the ones from plan, the order from state will be preserved"
}

func (m *keepIPAddressOrderPlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m *keepIPAddressOrderPlanModifier) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	if req.StateValue.IsNull() {
		return
	}

	var stateValues, planValues []string
	resp.Diagnostics.Append(req.StateValue.ElementsAs(ctx, &stateValues, true)...)
	resp.Diagnostics.Append(req.PlanValue.ElementsAs(ctx, &planValues, true)...)

	if cmp.Diff(stateValues, planValues, cmpopts.SortSlices(func(a, b string) bool { return a < b })) == "" {
		resp.PlanValue = req.StateValue
	}
}
