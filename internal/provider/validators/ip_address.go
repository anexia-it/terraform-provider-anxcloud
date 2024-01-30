package validators

import (
	"context"
	"net/netip"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = ipAddressValidator{}

type ipAddressValidator struct{}

func (v ipAddressValidator) Description(ctx context.Context) string {
	return "value must be a valid ip address; identifiers are no longer supported"
}

func (v ipAddressValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v ipAddressValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	val := req.ConfigValue.ValueString()

	if _, err := netip.ParseAddr(val); err != nil {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeValueMatchDiagnostic(
			req.Path,
			v.Description(ctx),
			val,
		))
	}
}

func ValidIPAddress() validator.String {
	return ipAddressValidator{}
}
