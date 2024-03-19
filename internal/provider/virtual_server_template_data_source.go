package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.anx.io/go-anxcloud/pkg/api"
	corev1 "go.anx.io/go-anxcloud/pkg/apis/core/v1"
	vspherev1 "go.anx.io/go-anxcloud/pkg/apis/vsphere/v1"
)

var _ datasource.DataSource = &VirtualServerTemplateDataSource{}
var _ datasource.DataSourceWithConfigure = &VirtualServerTemplateDataSource{}
var _ datasource.DataSourceWithConfigValidators = &VirtualServerTemplateDataSource{}

func NewVirtuaServerTemplateDataSource() datasource.DataSource {
	return &VirtualServerTemplateDataSource{}
}

type VirtualServerTemplateDataSource struct {
	engine api.API
}

func (*VirtualServerTemplateDataSource) ConfigValidators(context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.ExactlyOneOf(
			path.MatchRoot("id"),
			path.MatchRoot("name"),
		),
		datasourcevalidator.Conflicting(
			path.MatchRoot("id"),
			path.MatchRoot("build"),
		),
	}
}

type VirtualServerTemplateDataSourceModel struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Build    types.String `tfsdk:"build"`
	Location types.String `tfsdk:"location"`
}

func (ds *VirtualServerTemplateDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	ds.engine = req.ProviderData.(providerConfiguration).engine
}

func (ds *VirtualServerTemplateDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_virtual_server_template"
}

func (ds *VirtualServerTemplateDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves a virtual server template. Can be used to resolve a template ID by name, which is needed for creating anxcloud_virtual_server resources. " +
			"This datasource does not support 'from_scratch' templates!",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Datacenter location identifier.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Template name.",
				Optional:            true,
				Computed:            true,
			},
			"build": schema.StringAttribute{
				MarkdownDescription: "Template build.",
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

func (ds *VirtualServerTemplateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data VirtualServerTemplateDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	template := vspherev1.Template{
		Identifier: data.ID.ValueString(),
		Type:       vspherev1.TypeTemplate,
		Location:   corev1.Location{Identifier: data.Location.ValueString()},
	}

	if !data.ID.IsNull() {
		if err := ds.engine.Get(ctx, &template); err != nil {
			resp.Diagnostics.AddError("Could not find named template", err.Error())
			return
		}
	} else {
		tpl, err := vspherev1.FindNamedTemplate(
			ctx,
			ds.engine,
			data.Name.ValueString(),
			data.Build.ValueString(),
			corev1.Location{Identifier: data.Location.ValueString()},
		)
		if err != nil {
			resp.Diagnostics.AddError("Could not find named template", err.Error())
			return
		}

		template = *tpl
	}

	data = VirtualServerTemplateDataSourceModel{
		ID:       types.StringValue(template.Identifier),
		Name:     types.StringValue(template.Name),
		Build:    types.StringValue(template.Build),
		Location: data.Location,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
