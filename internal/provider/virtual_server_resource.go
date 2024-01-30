package provider

import (
	"context"
	"time"

	"github.com/anexia-it/terraform-provider-anxcloud/internal/provider/customtypes"
	"github.com/anexia-it/terraform-provider-anxcloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"go.anx.io/go-anxcloud/pkg/api"
	"go.anx.io/go-anxcloud/pkg/ipam"
	"go.anx.io/go-anxcloud/pkg/ipam/address"
	"go.anx.io/go-anxcloud/pkg/vsphere"
	"go.anx.io/go-anxcloud/pkg/vsphere/provisioning/nictype"
	"go.anx.io/go-anxcloud/pkg/vsphere/provisioning/vm"
)

var _ resource.Resource = &VirtualServerResource{}
var _ resource.ResourceWithImportState = &VirtualServerResource{}
var _ resource.ResourceWithConfigValidators = &VirtualServerResource{}

func NewVirtuaServerResource() resource.Resource {
	return &VirtualServerResource{}
}

type VirtualServerResource struct {
	engine     api.API
	vsphereAPI vsphere.API
	ipamAPI    ipam.API
	nicTypeAPI nictype.API
}

type VirtualServerResourceModel struct {
	ID                         types.String                              `tfsdk:"id"`
	Hostname                   customtypes.HostnameStringValue           `tfsdk:"hostname"`
	Location                   types.String                              `tfsdk:"location_id"`
	Template                   types.String                              `tfsdk:"template_id"`
	TemplateType               types.String                              `tfsdk:"template_type"`
	CPUs                       types.Int64                               `tfsdk:"cpus"`
	CPUPerformanceType         customtypes.CPUPerformanceTypeStringValue `tfsdk:"cpu_performance_type"`
	CPUSockets                 types.Int64                               `tfsdk:"sockets"`
	Memory                     types.Int64                               `tfsdk:"memory"`
	Disks                      types.List                                `tfsdk:"disk"`
	Networks                   types.List                                `tfsdk:"network"`
	DNS                        types.List                                `tfsdk:"dns"`
	Password                   types.String                              `tfsdk:"password"`
	SSH                        types.String                              `tfsdk:"ssh_key"`
	Script                     types.String                              `tfsdk:"script"`
	BootDelay                  types.Int64                               `tfsdk:"boot_delay"`
	EnterBIOSSetup             types.Bool                                `tfsdk:"enter_bios_setup"`
	ForceRestartIfNeeded       types.Bool                                `tfsdk:"force_restart_if_needed"`
	CriticalOperationConfirmed types.Bool                                `tfsdk:"critical_operation_confirmed"`

	Tags types.Set `tfsdk:"tags"`
}

type VirtualServerDiskModel struct {
	ID     types.Int64  `tfsdk:"disk_id"`
	SizeGB types.Int64  `tfsdk:"disk_gb"`
	Type   types.String `tfsdk:"disk_type"`
}

type VirtualServerNetworkModel struct {
	VLAN    types.String `tfsdk:"vlan_id"`
	NICType types.String `tfsdk:"nic_type"`
	IPs     types.List   `tfsdk:"ips"`
}

func (r *VirtualServerResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_virtual_server"
}

func (r *VirtualServerResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerConfig := req.ProviderData.(providerConfiguration)
	r.vsphereAPI = vsphere.NewAPI(providerConfig.legacyClient)
	r.ipamAPI = ipam.NewAPI(providerConfig.legacyClient)
	r.nicTypeAPI = nictype.NewAPI(providerConfig.legacyClient)
	r.engine = providerConfig.engine
}

func (*VirtualServerResource) ConfigValidators(context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("password"),
			path.MatchRoot("ssh_key"),
		),
	}
}

func (r *VirtualServerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VirtualServerResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var dns [4]string
	var dnsFromPlan []string
	resp.Diagnostics.Append(data.DNS.ElementsAs(ctx, &dnsFromPlan, false)...)
	for i := 0; i < len(dnsFromPlan) && i < len(dns); i++ {
		dns[i] = dnsFromPlan[i]
	}

	var planDisks []VirtualServerDiskModel
	resp.Diagnostics.Append(data.Disks.ElementsAs(ctx, &planDisks, false)...)
	disks := make([]vm.AdditionalDisk, 0, len(planDisks))
	for _, disk := range planDisks {
		disks = append(disks, vm.AdditionalDisk{
			SizeGBs: int(disk.SizeGB.ValueInt64()),
			Type:    disk.Type.ValueString(),
		})
	}

	var planNetworks []VirtualServerNetworkModel
	resp.Diagnostics.Append(data.Networks.ElementsAs(ctx, &planNetworks, false)...)
	networks := make([]vm.Network, 0, len(planNetworks))
	for _, network := range planNetworks {
		var ips []string
		resp.Diagnostics.Append(network.IPs.ElementsAs(ctx, &ips, true)...)

		if len(ips) == 0 {
			reserveSummary, err := r.ipamAPI.Address().ReserveRandom(ctx, address.ReserveRandom{
				LocationID: data.Location.ValueString(),
				VlanID:     network.VLAN.ValueString(),
				Count:      1,
			})
			if err != nil {
				resp.Diagnostics.AddError("Unable to reserve random address", err.Error())
				return
			}

			ips = append(ips, reserveSummary.Data[0].Address)
		}

		networks = append(networks, vm.Network{
			VLAN:    network.VLAN.ValueString(),
			NICType: network.NICType.ValueString(),
			IPs:     ips,
		})
	}

	if resp.Diagnostics.HasError() {
		return
	}

	create := vm.Definition{
		Hostname:           data.Hostname.ValueString(),
		Location:           data.Location.ValueString(),
		TemplateID:         data.Template.ValueString(),
		TemplateType:       data.TemplateType.ValueString(),
		Memory:             int(data.Memory.ValueInt64()),
		CPUs:               int(data.CPUs.ValueInt64()),
		CPUPerformanceType: data.CPUPerformanceType.ValueString(),
		Sockets:            int(data.CPUSockets.ValueInt64()),
		Disk:               disks[0].SizeGBs,
		DiskType:           disks[0].Type,
		AdditionalDisks:    disks[1:],
		Network:            networks,
		DNS1:               dns[0],
		DNS2:               dns[1],
		DNS3:               dns[2],
		DNS4:               dns[3],
		Password:           data.Password.ValueString(),
		SSH:                data.SSH.ValueString(),
		Script:             data.Script.ValueString(),
		BootDelay:          int(data.BootDelay.ValueInt64()),
		EnterBIOSSetup:     data.EnterBIOSSetup.ValueBool(),
	}

	provisioning, err := r.vsphereAPI.Provisioning().VM().Provision(ctx, create, true)
	if err != nil {
		resp.Diagnostics.AddError("failed provisioning vm", err.Error())
		return
	}

	vmIdentifier, err := r.vsphereAPI.Provisioning().Progress().AwaitCompletion(ctx, provisioning.Identifier)
	if err != nil {
		resp.Diagnostics.AddError("failed awaiting vm provisioning", err.Error())
		return
	}

	data.ID = types.StringValue(vmIdentifier)

	resp.Diagnostics.Append(ensureTags(ctx, r.engine, vmIdentifier, req.Plan)...)

	time.Sleep(2 * time.Minute) // need to wait for guest tools to report data

	if diags, notFound := r.setFromInfo(ctx, &data); notFound {
		resp.State.RemoveResource(ctx)
		return
	} else {
		resp.Diagnostics.Append(diags...)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VirtualServerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state VirtualServerResourceModel

	resp.Diagnostics.Append(resp.State.Get(ctx, &state)...)

	if diags, notFound := r.setFromInfo(ctx, &state); notFound {
		resp.State.RemoveResource(ctx)
		return
	} else {
		resp.Diagnostics.Append(diags...)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *VirtualServerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state VirtualServerResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	change := vm.Change{
		Reboot:          plan.ForceRestartIfNeeded.ValueBool(),
		EnableDangerous: plan.CriticalOperationConfirmed.ValueBool(),
	}

	if !plan.Tags.Equal(state.Tags) {
		resp.Diagnostics.Append(ensureTags(ctx, r.engine, state.ID.ValueString(), req.Plan)...)
	}

	needsUpdate := false
	if !plan.Memory.Equal(state.Memory) {
		needsUpdate = true
		change.MemoryMBs = int(plan.Memory.ValueInt64())
	}
	if !plan.CPUs.Equal(state.CPUs) {
		needsUpdate = true
		change.CPUs = int(plan.CPUs.ValueInt64())
	}
	if !plan.CPUSockets.Equal(state.CPUSockets) {
		needsUpdate = true
		change.CPUSockets = int(plan.CPUSockets.ValueInt64())
	}
	if !plan.CPUPerformanceType.Equal(state.CPUPerformanceType) {
		needsUpdate = true
		change.CPUPerformanceType = plan.CPUPerformanceType.ValueString()
	}
	if !plan.BootDelay.Equal(state.BootDelay) {
		needsUpdate = true
		change.BootDelaySecs = int(plan.BootDelay.ValueInt64())
	}
	if !plan.EnterBIOSSetup.Equal(state.EnterBIOSSetup) {
		needsUpdate = true
		change.EnterBIOSSetup = plan.EnterBIOSSetup.ValueBool()
	}
	if !plan.Disks.Equal(state.Disks) {
		needsUpdate = true
		var disksFromPlan, disksFromState []VirtualServerDiskModel
		resp.Diagnostics.Append(plan.Disks.ElementsAs(ctx, &disksFromPlan, false)...)
		resp.Diagnostics.Append(state.Disks.ElementsAs(ctx, &disksFromState, false)...)

		for _, diskFromState := range disksFromState {
			diskInPlan := false
			for _, diskFromPlan := range disksFromPlan {
				if diskFromPlan.ID.Equal(diskFromState.ID) {
					diskInPlan = true

					if !diskFromPlan.Type.Equal(diskFromState.Type) ||
						!diskFromPlan.SizeGB.Equal(diskFromState.SizeGB) {
						change.ChangeDisks = append(change.ChangeDisks, vm.Disk{
							ID:      int(diskFromPlan.ID.ValueInt64()),
							Type:    diskFromPlan.Type.ValueString(),
							SizeGBs: int(diskFromPlan.SizeGB.ValueInt64()),
						})
					}
				}
			}

			if !diskInPlan {
				change.DeleteDiskIDs = append(change.DeleteDiskIDs, int(diskFromState.ID.ValueInt64()))
			}
		}

		for _, diskFromPlan := range disksFromPlan {
			if diskFromPlan.ID.IsUnknown() {
				change.AddDisks = append(change.AddDisks, vm.Disk{
					Type:    diskFromPlan.Type.ValueString(),
					SizeGBs: int(diskFromPlan.SizeGB.ValueInt64()),
				})
			}
		}
	}

	if needsUpdate {
		provisioning, err := r.vsphereAPI.Provisioning().VM().Update(ctx, state.ID.ValueString(), change)
		if err != nil {
			resp.Diagnostics.AddError("error updating vm", err.Error())
			return
		}

		if _, err = r.vsphereAPI.Provisioning().Progress().AwaitCompletion(ctx, provisioning.Identifier); err != nil {
			resp.Diagnostics.AddError("error waiting for vm update to complete", err.Error())
			return
		}

		time.Sleep(time.Minute) // need to wait for guest tools to report data
	}

	if diags, notFound := r.setFromInfo(ctx, &plan); notFound {
		resp.State.RemoveResource(ctx)
		return
	} else {
		resp.Diagnostics.Append(diags...)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *VirtualServerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state VirtualServerResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	deprovisioning, err := r.vsphereAPI.Provisioning().VM().Deprovision(ctx, state.ID.ValueString(), false)
	if utils.IsLegacyClientNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	} else if err != nil {
		resp.Diagnostics.AddError("error deleting vm", err.Error())
		return
	}

	_, err = r.vsphereAPI.Provisioning().Progress().AwaitCompletion(ctx, deprovisioning.Identifier)
	if err != nil {
		resp.Diagnostics.AddError("error awaiting vm deletion", err.Error())
	}
}

func (r *VirtualServerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	info, err := r.vsphereAPI.Info().Get(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to fetch virtual server", err.Error())
		return
	}

	if info.TemplateID == "" {
		resp.Diagnostics.AddError(
			"Cannot import virtual server with `from_scratch` template",
			"Importing virtual servers which have been provisioned with a `from_scratch` template "+
				"is not supported.",
		)
		return
	}

	resp.Diagnostics.AddWarning(
		"Resource Import Considerations",
		"Virtual server import does not include 'password' and 'ssh_key' attributes. "+
			"To prevent the virtual server from getting replaced in the next apply, make sure to add "+
			"either 'password' or 'ssh_key' (depending on which attribute is configured) to the 'ignore_changes' attribute "+
			"in the lifecycle block.",
	)

	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
