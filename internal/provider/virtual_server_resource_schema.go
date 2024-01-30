package provider

import (
	"context"

	"github.com/anexia-it/terraform-provider-anxcloud/internal/provider/customtypes"
	"github.com/anexia-it/terraform-provider-anxcloud/internal/provider/planmodifiers"
	"github.com/anexia-it/terraform-provider-anxcloud/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *VirtualServerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
The virtual_server resource allows you to configure and run virtual machines.

### Known limitations
- removal of disks not supported
- removal of networks not supported
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Virtual server identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"hostname": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Virtual server hostname",
				CustomType:          customtypes.HostnameStringType{},
				PlanModifiers: []planmodifier.String{
					planmodifiers.KeepStringPrefix(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"location_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Location identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"template_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Template identifier",
			},
			"template_type": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "OS template type",
				Default:             stringdefault.StaticString("templates"),
				Validators: []validator.String{
					stringvalidator.OneOf("templates", "from_scratch"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"cpus": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "Number of CPUs",
			},
			"cpu_performance_type": schema.StringAttribute{
				CustomType:          customtypes.CPUPerformanceTypeStringType{},
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "CPU type. Example: (`best-effort`, `standard`, `enterprise`, `performance`), defaults to `standard`.",
				Default:             stringdefault.StaticString("standard"),
				PlanModifiers: []planmodifier.String{
					planmodifiers.KeepStringSuffix(),
				},
			},
			"sockets": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				MarkdownDescription: "Amount of CPU sockets Number of cores have to be a multiple of sockets, as they will be spread evenly across all sockets. " +
					"Defaults to number of cores, i.e. one socket per CPU core.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"memory": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "Memory in MB.",
			},
			"dns": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "DNS configuration. Maximum items 4. Defaults to template settings.",
				Validators: []validator.List{
					listvalidator.SizeAtMost(4),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"password": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Plaintext password. Example: ('!anx123mySuperStrongPassword123anx!', 'go3ju0la1ro3', …). For systems that support it, we strongly recommend using a SSH key instead.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"ssh_key": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Public key (instead of password, only for Linux systems). Recommended over providing a plaintext password.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.Expressions{
						path.MatchRoot("password"),
					}...),
				},
			},
			"script": schema.StringAttribute{
				Optional: true,
				MarkdownDescription: "Script to be executed after provisioning. " +
					"Consider the corresponding shebang at the beginning of your script. " +
					"If you want to use PowerShell, the first line should be: #ps1_sysnative.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"boot_delay": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Boot delay in seconds. Example: (0, 1, …).",
			},
			"enter_bios_setup": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Start the VM into BIOS setup on next boot.",
			},
			"force_restart_if_needed": schema.BoolAttribute{
				Optional: true,
				MarkdownDescription: "Certain operations may only be performed in powered off state. " +
					"Such as: shrinking memory, shrinking/adding CPU, removing disk and scaling a disk beyond 2 GB. " +
					"Passing this value as true will always execute a power off and reboot request after completing all other operations. " +
					"Without this flag set to true scaling operations requiring a reboot will fail.",
			},
			"critical_operation_confirmed": schema.BoolAttribute{
				Optional: true,
				MarkdownDescription: "Confirms a critical operation (if needed). " +
					"Potentially dangerous operations (e.g. resulting in data loss) require an additional confirmation. " +
					"The parameter is used for VM UPDATE requests.",
			},

			"tags": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Set of tags attached to the resource",
			},
		},
		Blocks: map[string]schema.Block{
			"disk":    r.disksSchema(),
			"network": r.networksSchema(),
		},
	}
}

func (r *VirtualServerResource) disksSchema() schema.Block {
	return schema.ListNestedBlock{
		MarkdownDescription: "Virtual Server Disk.",
		Validators: []validator.List{
			listvalidator.IsRequired(),
			listvalidator.SizeAtLeast(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"disk_id": schema.Int64Attribute{
					Computed:            true,
					MarkdownDescription: "Device identifier of the disk.",
					PlanModifiers: []planmodifier.Int64{
						int64planmodifier.UseStateForUnknown(),
					},
				},
				"disk_gb": schema.Int64Attribute{
					Required:            true,
					MarkdownDescription: "Disk capacity in GB.",
				},
				"disk_type": schema.StringAttribute{
					Required:            true,
					MarkdownDescription: "Disk category (limits disk performance, e.g. IOPS).",
				},
			},
		},
	}
}

func (r *VirtualServerResource) networksSchema() schema.Block {
	return schema.ListNestedBlock{
		MarkdownDescription: "Network interface.",
		Validators: []validator.List{
			listvalidator.IsRequired(),
			listvalidator.SizeAtLeast(1),
		},
		PlanModifiers: []planmodifier.List{
			listplanmodifier.RequiresReplaceIf(func(ctx context.Context, req planmodifier.ListRequest, resp *listplanmodifier.RequiresReplaceIfFuncResponse) {
				if req.State.Raw.IsNull() {
					return
				}

				var plan, state []VirtualServerNetworkModel
				resp.Diagnostics.Append(req.PlanValue.ElementsAs(ctx, &plan, false)...)
				resp.Diagnostics.Append(req.StateValue.ElementsAs(ctx, &state, false)...)

				if resp.Diagnostics.HasError() {
					return
				}

				resp.RequiresReplace = len(state) != len(plan)
			}, "", ""),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"vlan_id": schema.StringAttribute{
					Required:            true,
					MarkdownDescription: "VLAN identifier.",
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"nic_type": schema.StringAttribute{
					Required:    true,
					Description: "Network interface card type.",
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"ips": schema.ListAttribute{
					Optional:    true,
					Computed:    true,
					ElementType: types.StringType,
					Validators: []validator.List{
						listvalidator.ValueStringsAre(validators.ValidIPAddress()),
						listvalidator.UniqueValues(),
					},
					PlanModifiers: []planmodifier.List{
						listplanmodifier.UseStateForUnknown(),
						planmodifiers.KeepIPAddressOrderPlanModifier(),
						listplanmodifier.RequiresReplaceIfConfigured(),
					},
					MarkdownDescription: "List of IP addresses and identifiers to be assigned and configured.",
				},
			},
		},
	}
}
