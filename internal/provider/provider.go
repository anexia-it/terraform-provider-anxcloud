package provider

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/anexia-it/terraform-provider-anxcloud/anxcloud"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.anx.io/go-anxcloud/pkg/client"
)

var _ provider.Provider = &AnexiaProvider{}

type AnexiaProvider struct {
	version string
}

type AnexiaProviderModel struct {
	Token types.String `tfsdk:"token"`
}

func (p *AnexiaProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "anxcloud"
	resp.Version = p.version
}

func (p *AnexiaProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"token": schema.StringAttribute{
				Optional:    true,
				Description: "Anexia Cloud token.",
				Sensitive:   true,
			},
		},
	}
}

func (p *AnexiaProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	apiToken := os.Getenv("ANEXIA_TOKEN")

	var data AnexiaProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.Token.ValueString() != "" {
		apiToken = data.Token.ValueString()
	}

	if apiToken == "" {
		resp.Diagnostics.AddError("Missing Anexia Engine token", "The Anexia Engine token was neither configured via `ANEXIA_TOKEN` env var nor via provider argument")
	}

	logger := anxcloud.NewTerraformr(log.Default().Writer())
	opts := []client.Option{
		client.TokenFromString(apiToken),
		client.Logger(logger.WithName("client")),
		client.UserAgent(fmt.Sprintf("%s/%s (%s)", "terraform-provider-anxcloud", p.version, runtime.GOOS)),
	}

	resp.ResourceData = opts
	resp.DataSourceData = opts
}

func (p *AnexiaProvider) Resources(ctx context.Context) []func() resource.Resource {
	return nil
}

func (p *AnexiaProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return nil
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &AnexiaProvider{
			version: version,
		}
	}
}
