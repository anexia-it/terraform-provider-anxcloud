package provider

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/anexia-it/terraform-provider-anxcloud/internal/provider/customtypes"
	"github.com/anexia-it/terraform-provider-anxcloud/internal/utils"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.anx.io/go-anxcloud/pkg/vsphere/info"
	"go.anx.io/go-anxcloud/pkg/vsphere/provisioning/nictype"
)

func (r *VirtualServerResource) setFromInfo(ctx context.Context, data *VirtualServerResourceModel) (diags diag.Diagnostics, notFound bool) {
	info, err := r.vsphereAPI.Info().Get(ctx, data.ID.ValueString())
	if utils.IsLegacyClientNotFound(err) {
		return nil, true
	} else if err != nil {
		diags.AddError("failed reading vm info", err.Error())
		return diags, false
	}

	var (
		templateType types.String

		template = data.Template
		networks = data.Networks
	)

	// if template type == "templates" -> use data from info endpoint
	// else -> use local data
	if info.TemplateID != "" {
		template = types.StringValue(info.TemplateID)
		templateType = types.StringValue("templates")
		diags.Append(r.toNetworkList(ctx, info.Network, &networks)...)
	} else {
		templateType = types.StringValue("from_scratch")
	}

	*data = VirtualServerResourceModel{
		ID:                 types.StringValue(info.Identifier),
		Hostname:           customtypes.HostnameValue(info.Name),
		Template:           template,
		TemplateType:       templateType,
		Location:           types.StringValue(info.LocationID),
		CPUs:               types.Int64Value(int64(info.CPU)),
		CPUPerformanceType: customtypes.CPUPerformanceTypeValue(info.CPUPerformanceType),
		CPUSockets:         types.Int64Value(int64(info.CPU / info.Cores)),
		Memory:             types.Int64Value(int64(info.RAM)),
		Networks:           networks,

		// not returned by API -> take over from state
		DNS:                        data.DNS,
		Password:                   data.Password,
		SSH:                        data.SSH,
		Script:                     data.Script,
		BootDelay:                  data.BootDelay,
		EnterBIOSSetup:             data.EnterBIOSSetup,
		CriticalOperationConfirmed: data.CriticalOperationConfirmed,
		ForceRestartIfNeeded:       data.ForceRestartIfNeeded,
	}

	diags.Append(readTags(ctx, r.engine, info.Identifier, &data.Tags)...)

	diags.Append(r.toDiskList(ctx, info.DiskInfo, &data.Disks)...)

	return diags, false
}

func (r *VirtualServerResource) toDiskList(ctx context.Context, infoDisks []info.DiskInfo, list *types.List) (diags diag.Diagnostics) {
	var listDisks []VirtualServerDiskModel

	for _, disk := range infoDisks {
		listDisks = append(listDisks, VirtualServerDiskModel{
			ID:     types.Int64Value(int64(disk.DiskID)),
			SizeGB: types.Int64Value(int64(math.Round(disk.DiskGB))),
			Type:   types.StringValue(disk.DiskType),
		})
	}

	*list, diags = types.ListValueFrom(ctx, r.disksSchema().GetNestedObject().Type(), listDisks)
	return diags
}

func (r *VirtualServerResource) toNetworkList(ctx context.Context, infoNetworks []info.Network, list *types.List) (diags diag.Diagnostics) {
	var prevNetworkList []VirtualServerNetworkModel
	diags.Append(list.ElementsAs(ctx, &prevNetworkList, false)...)

	var networkList []VirtualServerNetworkModel

	for i, network := range infoNetworks {
		nicType, err := nicTypeFromID(ctx, r.nicTypeAPI, network.NIC)
		if err != nil {
			diags.AddError("unknown nic type", err.Error())
		}

		ips := append(network.IPv4, network.IPv6...)
		// order is not stable -> if previous state contains same elements, use that to prevent inconsitency errors
		if len(prevNetworkList) > i {
			var prevIPs []string
			diags.Append(prevNetworkList[i].IPs.ElementsAs(ctx, &prevIPs, true)...)
			if cmp.Diff(ips, prevIPs, cmpopts.SortSlices(func(a, b string) bool { return a < b })) == "" {
				ips = prevIPs
			}
		}

		ipList, ipListDiags := types.ListValueFrom(ctx, types.StringType, ips)
		diags.Append(ipListDiags...)

		networkList = append(networkList, VirtualServerNetworkModel{
			VLAN:    types.StringValue(network.VLAN),
			NICType: types.StringValue(nicType),
			IPs:     ipList,
		})
	}

	if diags.HasError() {
		return diags
	}

	*list, diags = types.ListValueFrom(ctx, r.networksSchema().GetNestedObject().Type(), networkList)
	return diags
}

func nicTypeFromID(ctx context.Context, nicTypeAPI nictype.API, nicTypeID int) (string, error) {
	nicTypeIndex := nicTypeID - 1

	types, err := nicTypeAPI.List(ctx)
	if err != nil {
		return "", fmt.Errorf("fetch available nic types: %w", err)
	}

	if nicTypeIndex < 0 || nicTypeIndex >= len(types) {
		return "", errors.New("nic type not found")
	}

	return types[nicTypeIndex], nil
}
