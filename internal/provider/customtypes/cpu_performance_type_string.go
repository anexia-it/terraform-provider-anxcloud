package customtypes

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var _ basetypes.StringValuable = CPUPerformanceTypeStringValue{}
var _ basetypes.StringValuableWithSemanticEquals = CPUPerformanceTypeStringValue{}

type CPUPerformanceTypeStringValue struct {
	basetypes.StringValue
}

func (v CPUPerformanceTypeStringValue) StringSemanticEquals(ctx context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(CPUPerformanceTypeStringValue)
	if !ok {
		diags.AddError(
			"Semantic Equality Check Error",
			"Received unexpected value type",
		)

		return false, diags
	}

	return strings.HasPrefix(v.ValueString(), newValue.ValueString()), diags
}

func CPUPerformanceTypeValue(value string) CPUPerformanceTypeStringValue {
	return CPUPerformanceTypeStringValue{
		StringValue: types.StringValue(value),
	}
}

var _ basetypes.StringTypable = CPUPerformanceTypeStringType{}

type CPUPerformanceTypeStringType struct {
	basetypes.StringType
}

func (t CPUPerformanceTypeStringType) String() string {
	return "CPUPerformanceTypeStringType"
}

func (t CPUPerformanceTypeStringType) ValueFromString(ctx context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	value := CPUPerformanceTypeStringValue{
		StringValue: in,
	}

	return value, nil
}

func (t CPUPerformanceTypeStringType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.StringType.ValueFromTerraform(ctx, in)

	if err != nil {
		return nil, err
	}

	stringValue, ok := attrValue.(basetypes.StringValue)

	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	stringValuable, diags := t.ValueFromString(ctx, stringValue)

	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting StringValue to StringValuable: %v", diags)
	}

	return stringValuable, nil
}

func (t CPUPerformanceTypeStringType) ValueType(ctx context.Context) attr.Value {
	return CPUPerformanceTypeStringValue{}
}
