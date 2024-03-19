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

var _ basetypes.StringValuable = HostnameStringValue{}
var _ basetypes.StringValuableWithSemanticEquals = HostnameStringValue{}

type HostnameStringValue struct {
	basetypes.StringValue
}

func (v HostnameStringValue) StringSemanticEquals(ctx context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(HostnameStringValue)
	if !ok {
		diags.AddError(
			"Semantic Equality Check Error",
			"Received unexpected value type",
		)

		return false, diags
	}

	return strings.HasSuffix(v.ValueString(), newValue.ValueString()), diags
}

func HostnameValue(value string) HostnameStringValue {
	return HostnameStringValue{
		StringValue: types.StringValue(value),
	}
}

var _ basetypes.StringTypable = HostnameStringType{}

type HostnameStringType struct {
	basetypes.StringType
}

func (t HostnameStringType) String() string {
	return "HostnameStringType"
}

func (t HostnameStringType) ValueFromString(ctx context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	value := HostnameStringValue{
		StringValue: in,
	}

	return value, nil
}

func (t HostnameStringType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
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

func (t HostnameStringType) ValueType(ctx context.Context) attr.Value {
	return HostnameStringValue{}
}
