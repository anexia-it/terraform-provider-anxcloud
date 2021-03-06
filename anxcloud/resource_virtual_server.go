package anxcloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-multierror"
	"log"
	"time"

	"github.com/anexia-it/go-anxcloud/pkg/client"
	"github.com/anexia-it/go-anxcloud/pkg/ipam/address"
	"github.com/anexia-it/go-anxcloud/pkg/vsphere"
	"github.com/anexia-it/go-anxcloud/pkg/vsphere/provisioning/nictype"
	"github.com/anexia-it/go-anxcloud/pkg/vsphere/provisioning/vm"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	maxDNSEntries = 4
	vmPoweredOn   = "poweredOn"
	vmPoweredOff  = "poweredOff"
)

func resourceVirtualServer() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVirtualServerCreate,
		ReadContext:   resourceVirtualServerRead,
		UpdateContext: resourceVirtualServerUpdate,
		DeleteContext: resourceVirtualServerDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Read:   schema.DefaultTimeout(1 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: schemaVirtualServer(),
		CustomizeDiff: customdiff.All(
			customdiff.ForceNewIf("network", func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) bool {
				old, newNetworks := d.GetChange("network")
				oldNets := expandVirtualServerNetworks(old.([]interface{}))
				newNets := expandVirtualServerNetworks(newNetworks.([]interface{}))

				if len(oldNets) > len(newNets) {
					// some network has been deleted
					return true
				}

				// Get the IPs which are associated with the VM from info.network key
				vmInfoState := d.Get("info").([]interface{})
				infoObject := expandVirtualServerInfo(vmInfoState)
				vmIPMap := make(map[string]struct{})
				for _, vmNet := range infoObject.Network {
					for _, ip := range append(vmNet.IPv4, vmNet.IPv6...) {
						vmIPMap[ip] = struct{}{}
					}
				}

				for i, newNet := range newNets {
					if i+1 > len(oldNets) {
						// new networks were added
						break
					}

					if newNet.VLAN != oldNets[i].VLAN {
						key := fmt.Sprintf("network.%d.vlan_id", i)
						if err := d.ForceNew(key); err != nil {
							log.Fatalf("[ERROR] unable to force new '%s': %v", key, err)
						}
					}

					if newNet.NICType != oldNets[i].NICType {
						key := fmt.Sprintf("network.%d.nic_type", i)
						if err := d.ForceNew(key); err != nil {
							log.Fatalf("[ERROR] unable to force new '%s': %v", key, err)
						}
					}

					if len(newNet.IPs) < len(oldNets[i].IPs) {
						// IPs are missing
						key := fmt.Sprintf("network.%d.ips", i)
						if err := d.ForceNew(key); err != nil {
							log.Fatalf("[ERROR] unable to force new '%s': %v", key, err)
						}
					} else {
						for j, ip := range newNet.IPs {
							if j >= len(oldNets[i].IPs) || ip != oldNets[i].IPs[j] {
								if _, ipExpected := vmIPMap[ip]; ipExpected {
									continue
								}

								key := fmt.Sprintf("network.%d.ips", i)
								if err := d.ForceNew(key); err != nil {
									log.Fatalf("[ERROR] unable to force new '%s': %v", key, err)
								}
							}
						}
					}
				}

				return false
			}),
			customdiff.ForceNewIf("disk", func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) bool {
				old, new := d.GetChange("disk")
				oldDisks := expandVirtualServerDisks(old.([]interface{}))
				newDisks := expandVirtualServerDisks(new.([]interface{}))

				if len(oldDisks) > len(newDisks) {
					return true
				}

				for i, disk := range newDisks {
					if i+1 > len(oldDisks) {
						// new disks were added
						break
					}

					if disk.SizeGBs < oldDisks[i].SizeGBs {
						key := fmt.Sprintf("disk.%d.disk_gb", i)
						if err := d.ForceNew(key); err != nil {
							log.Fatalf("[ERROR] unable to force new '%s': %v", key, err)
						}
					}
				}
				return false
			}),
		),
	}
}

func resourceVirtualServerCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var (
		diags    diag.Diagnostics
		networks []vm.Network
		disks    []Disk
	)

	c := m.(client.Client)
	vsphereAPI := vsphere.NewAPI(c)
	addressAPI := address.NewAPI(c)
	locationID := d.Get("location_id").(string)

	networks = expandVirtualServerNetworks(d.Get("network").([]interface{}))
	for i, n := range networks {
		if len(n.IPs) > 0 {
			continue
		}

		res, err := addressAPI.ReserveRandom(ctx, address.ReserveRandom{
			LocationID: locationID,
			VlanID:     n.VLAN,
			Count:      1,
		})
		if err != nil {
			diags = append(diags, diag.FromErr(err)...)
		} else if len(res.Data) > 0 {
			networks[i].IPs = append(networks[i].IPs, res.Data[0].Address)
		} else {
			diags = append(diags, diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       "Free IP not found",
				Detail:        fmt.Sprintf("Free IP not found for VLAN: '%s'", n.VLAN),
				AttributePath: cty.Path{cty.GetAttrStep{Name: "ips"}},
			})
		}
	}

	dns := expandVirtualServerDNS(d.Get("dns").([]interface{}))
	if len(dns) != maxDNSEntries {
		diags = append(diags, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       "DNS entries exceed limit",
			Detail:        "Number of DNS entries cannot exceed limit 4",
			AttributePath: cty.Path{cty.GetAttrStep{Name: "dns"}},
		})
	}

	disks = expandVirtualServerDisks(d.Get("disk").([]interface{}))

	// We require at least one disk to be specified either via Disk or via Disks array
	if len(disks) < 1 {
		diags = append(diags, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       "No disk specified",
			Detail:        "Minimum of one disk has to be specified",
			AttributePath: cty.Path{cty.GetAttrStep{Name: "size_gb"}},
		})
	}

	if len(diags) > 0 {
		return diags
	}

	def := vm.Definition{
		Location:           locationID,
		TemplateType:       d.Get("template_type").(string),
		TemplateID:         d.Get("template_id").(string),
		Hostname:           d.Get("hostname").(string),
		Memory:             d.Get("memory").(int),
		CPUs:               d.Get("cpus").(int),
		Disk:               disks[0].SizeGBs, //Workaround until Create API supports multi disk
		DiskType:           disks[0].Type,
		CPUPerformanceType: d.Get("cpu_performance_type").(string),
		Sockets:            d.Get("sockets").(int),
		Network:            networks,
		DNS1:               dns[0],
		DNS2:               dns[1],
		DNS3:               dns[2],
		DNS4:               dns[3],
		Password:           d.Get("password").(string),
		SSH:                d.Get("ssh_key").(string),
		Script:             d.Get("script").(string),
		BootDelay:          d.Get("boot_delay").(int),
		EnterBIOSSetup:     d.Get("enter_bios_setup").(bool),
	}

	base64Encoding := true
	provision, err := vsphereAPI.Provisioning().VM().Provision(ctx, def, base64Encoding)
	if err != nil {
		return diag.FromErr(err)
	}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		if d.Id() == "" {
			p, err := vsphereAPI.Provisioning().Progress().Get(ctx, provision.Identifier)
			if err != nil {
				return resource.NonRetryableError(fmt.Errorf("unable to get vm progress by ID '%s', %w", provision.Identifier, err))
			}
			if p.VMIdentifier != "" && p.Progress < 100 {
				d.SetId(p.VMIdentifier)
			} else {
				return resource.RetryableError(fmt.Errorf("vm with provisioning ID '%s' is not ready yet: %d %%", provision.Identifier, p.Progress))
			}
		}

		vmInfo, err := vsphereAPI.Info().Get(ctx, d.Id())
		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("unable to get vm  by ID '%s', %w", d.Id(), err))
		}
		if vmInfo.Status == vmPoweredOn {
			return nil
		}
		return resource.RetryableError(fmt.Errorf("vm with id '%s' is not %s yet: %s", d.Id(), vmPoweredOn, vmInfo.Status))
	})

	if err != nil {
		return diag.FromErr(err)
	}

	tags := expandTags(d.Get("tags").([]interface{}))
	for _, t := range tags {
		if err := attachTag(ctx, c, d.Id(), t); err != nil {
			return diag.FromErr(err)
		}
	}

	if len(disks) > 1 {
		if read := resourceVirtualServerRead(ctx, d, m); read.HasError() {
			return read
		}

		initialDisks := expandVirtualServerDisks(d.Get("disk").([]interface{}))
		if update := updateVirtualServerDisk(ctx, c, d.Id(), disks, initialDisks); update != nil {
			return update
		}
	}

	diags = resourceVirtualServerRead(ctx, d, m)
	return diags
}

func resourceVirtualServerRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(client.Client)
	vsphereAPI := vsphere.NewAPI(c)
	nicAPI := nictype.NewAPI(c)

	info, err := vsphereAPI.Info().Get(ctx, d.Id())
	if err != nil {
		if err := handleNotFoundError(err); err != nil {
			return diag.FromErr(err)
		}
		d.SetId("")
		return nil
	}

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		info, infoErr := vsphereAPI.Info().Get(ctx, d.Id())
		if infoErr != nil {
			if err := handleNotFoundError(infoErr); err != nil {
				return resource.NonRetryableError(fmt.Errorf("unable to get vm with id '%s': %w", d.Id(), err))
			}
			d.SetId("")
			return nil
		}
		var nicErr error
		for _, nic := range info.Network {
			if len(nic.IPv4) == 0 && len(nic.IPv6) == 0 {
				nicErr = multierror.Append(nicErr, fmt.Errorf("missing IPs for NIC"))
			}
		}
		if nicErr != nil {
			return resource.RetryableError(nicErr)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	nicTypes, err := nicAPI.List(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	// TODO: we miss information about:
	// * cpu_performance_type

	if err = d.Set("location_id", info.LocationID); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err = d.Set("template_id", info.TemplateID); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	//if err = d.Set("template_type", info.TemplateType); err != nil {
	//	diags = append(diags, diag.FromErr(err)...)
	//}
	if err = d.Set("cpus", info.CPU); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if v := d.Get("sockets").(int); v != 0 {
		// TODO: API fix: info.Cores should be info.Sockets, there is info.Cpus which is info.Cores
		if err = d.Set("sockets", info.Cores); err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
	}

	if err = d.Set("memory", info.RAM); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	disks := make([]Disk, len(info.DiskInfo))
	for i, diskInfo := range info.DiskInfo {
		diskGB := roundDiskSize(diskInfo.DiskGB)
		disks[i] = Disk{
			Disk: &vm.Disk{
				ID:      diskInfo.DiskID,
				Type:    diskInfo.DiskType,
				SizeGBs: diskGB,
			},
			ExactDiskSize: diskInfo.DiskGB,
		}
	}

	flattenedDisks := flattenVirtualServerDisks(disks)
	if err = d.Set("disk", flattenedDisks); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	specNetworks := expandVirtualServerNetworks(d.Get("network").([]interface{}))
	networks := make([]vm.Network, 0, len(info.Network))
	for i, net := range info.Network {
		if len(nicTypes) < net.NIC {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Requested invalid nic type",
				Detail:   fmt.Sprintf("NIC type index out of range, available %d, wanted %d", len(nicTypes), net.NIC),
			})
			continue
		}

		if len(specNetworks) > i {
			expectedIPMap := make(map[string]struct{}, len(specNetworks[i].IPs))
			for _, ip := range specNetworks[i].IPs {
				expectedIPMap[ip] = struct{}{}
			}

			network := vm.Network{
				NICType: nicTypes[net.NIC-1],
				VLAN:    net.VLAN,
			}

			for _, ipv4 := range net.IPv4 {
				if _, ok := expectedIPMap[ipv4]; ok {
					network.IPs = append(network.IPs, ipv4)
					delete(expectedIPMap, ipv4)
				}
			}

			for _, ipv6 := range net.IPv6 {
				if _, ok := expectedIPMap[ipv6]; ok {
					network.IPs = append(network.IPs, ipv6)
					delete(expectedIPMap, ipv6)
				}
			}

			for ip := range expectedIPMap {
				network.IPs = append(network.IPs, ip)
			}

			networks = append(networks, network)
		}
	}

	flattenedNetworks := flattenVirtualServerNetwork(networks)
	if err = d.Set("network", flattenedNetworks); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	flattenedInfo := flattenVirtualServerInfo(&info)
	if err = d.Set("info", flattenedInfo); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	return diags
}

func resourceVirtualServerUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.Client)
	vsphereAPI := vsphere.NewAPI(c)
	ch := vm.Change{
		Reboot:          d.Get("force_restart_if_needed").(bool),
		EnableDangerous: d.Get("critical_operation_confirmed").(bool),
	}
	requiresReboot := false

	if d.HasChanges("sockets", "memory", "cpus") {
		ch.CPUs = d.Get("cpus").(int)
		ch.CPUSockets = d.Get("sockets").(int)
		ch.MemoryMBs = d.Get("memory").(int)

		requiresReboot = true
	}

	// must stay in a separate condition as any endpoint doesn't return info about the current state
	// thus we lose control over expected and current states
	if d.HasChange("cpu_performance_type") {
		ch.CPUPerformanceType = d.Get("cpu_performance_type").(string)
	}

	if d.HasChange("network") {
		old, new := d.GetChange("network")
		oldNets := expandVirtualServerNetworks(old.([]interface{}))
		newNets := expandVirtualServerNetworks(new.([]interface{}))

		if len(oldNets) < len(newNets) {
			ch.AddNICs = newNets[len(oldNets):]
		} else {
			return diag.Errorf(
				"unsupported update operation, cannot remove network or update its parameters",
			)
		}
	}

	if d.HasChange("disk") {
		old, new := d.GetChange("disk")
		oldDisks := expandVirtualServerDisks(old.([]interface{}))
		newDisks := expandVirtualServerDisks(new.([]interface{}))

		if len(newDisks) < len(oldDisks) {
			return diag.Errorf("removing disks is not supported yet, expected at least %d, got %d", len(oldDisks), len(newDisks))
		}

		changeDisks := make([]vm.Disk, 0, len(oldDisks))
		addDisks := make([]vm.Disk, 0, len(newDisks))
		for i := range newDisks {
			if i >= len(oldDisks) {
				addDisks = append(addDisks, *newDisks[i].Disk)
				continue
			}

			actualDisk := oldDisks[i]
			expectedDisk := newDisks[i]

			// Compare the floating point disk size with the changed disk size from the configuration.
			// This ensures that scaling operations are not reliant on rounding the disk size to integers.
			if actualDisk.Type != expectedDisk.Type || actualDisk.ExactDiskSize < float64(expectedDisk.SizeGBs) {
				changeDisks = append(changeDisks, *expectedDisk.Disk)
			}
		}
		ch.ChangeDisks = changeDisks
		ch.AddDisks = addDisks
	}

	if d.HasChange("tags") {
		old, new := d.GetChange("tags")
		oldTags := expandTags(old.([]interface{}))
		newTags := expandTags(new.([]interface{}))

		dTags := getTagsDifferences(oldTags, newTags)
		cTags := getTagsDifferences(newTags, oldTags)

		for _, t := range dTags {
			if err := detachTag(ctx, c, d.Id(), t); err != nil {
				return diag.FromErr(err)
			}
		}

		for _, t := range cTags {
			if err := attachTag(ctx, c, d.Id(), t); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	var response vm.ProvisioningResponse
	provisioningAPI := vsphereAPI.Provisioning()

	var err error
	if response, err = provisioningAPI.VM().Update(ctx, d.Id(), ch); err != nil {
		return diag.FromErr(err)
	}

	if _, err = provisioningAPI.Progress().AwaitCompletion(ctx, response.Identifier); err != nil {
		return diag.FromErr(err)
	}

	delay := 10 * time.Second
	if requiresReboot {
		delay = 3 * time.Minute
	}

	vmState := resource.StateChangeConf{
		Delay:      delay,
		Timeout:    d.Timeout(schema.TimeoutUpdate),
		MinTimeout: 10 * time.Second,
		Pending: []string{
			vmPoweredOff,
		},
		Target: []string{
			vmPoweredOn,
		},
		Refresh: func() (interface{}, string, error) {
			info, err := vsphereAPI.Info().Get(ctx, d.Id())
			if err != nil {
				return "", "", err
			}
			return info, info.Status, nil
		},
	}
	if _, err = vmState.WaitForStateContext(ctx); err != nil {
		return diag.FromErr(err)
	}

	return resourceVirtualServerRead(ctx, d, m)
}

func resourceVirtualServerDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.Client)
	vsphereAPI := vsphere.NewAPI(c)

	delayedDeprovision := false
	err := vsphereAPI.Provisioning().VM().Deprovision(ctx, d.Id(), delayedDeprovision)
	if err != nil {
		if err := handleNotFoundError(err); err != nil {
			return diag.FromErr(err)
		}
		d.SetId("")
		return nil
	}

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err := vsphereAPI.Info().Get(ctx, d.Id())
		if err != nil {
			if err := handleNotFoundError(err); err != nil {
				return resource.NonRetryableError(fmt.Errorf("unable to get vm with id '%s': %w", d.Id(), err))
			}
			d.SetId("")
			return nil
		}
		return resource.RetryableError(fmt.Errorf("waiting for vm with id '%s' to be deleted", d.Id()))
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func getTagsDifferences(tagsA, tagsB []string) []string {
	var out []string

	for _, a := range tagsA {
		found := false
		for _, b := range tagsB {
			if a == b {
				found = true
			}
		}
		if !found {
			out = append(out, a)
		}
	}

	return out
}

func updateVirtualServerDisk(ctx context.Context, c client.Client, id string, expected []Disk, current []Disk) diag.Diagnostics {
	changeDisks := make([]vm.Disk, 0, len(current))
	addDisks := make([]vm.Disk, 0, len(expected))
	for diskIndex := range current {
		if diskIndex >= len(expected) {
			break
		}
		expected[diskIndex].ID = current[diskIndex].ID
		actualDisk := current[diskIndex]
		expectedDisk := expected[diskIndex]

		if actualDisk.ExactDiskSize > float64(expectedDisk.SizeGBs) {
			LogError("Skipping disk %d because expected disk size to small! Expected: %d  -  got: %f", actualDisk.ID, expectedDisk.SizeGBs, actualDisk.ExactDiskSize)
		}
		if actualDisk.Type != expectedDisk.Type || actualDisk.ExactDiskSize < float64(expectedDisk.SizeGBs) {
			changeDisks = append(changeDisks, *expectedDisk.Disk)
		}
	}

	if len(expected) > len(current) {
		for newDiskIndex := len(current); newDiskIndex < len(expected); newDiskIndex++ {
			addDisks = append(addDisks, *expected[newDiskIndex].Disk)
		}
	}

	ch := vm.Change{
		AddDisks:    addDisks,
		ChangeDisks: changeDisks,
	}

	v := vsphere.NewAPI(c)
	var response vm.ProvisioningResponse
	provisioning := v.Provisioning()
	var err error
	if response, err = provisioning.VM().Update(ctx, id, ch); err != nil {
		return diag.FromErr(err)
	}

	if _, err = provisioning.Progress().AwaitCompletion(ctx, response.Identifier); err != nil {
		return diag.FromErr(err)
	}

	vmState := resource.StateChangeConf{
		Delay:      10 * time.Second,
		Timeout:    10 * time.Minute,
		MinTimeout: 10 * time.Second,
		Pending: []string{
			vmPoweredOff,
		},
		Target: []string{
			vmPoweredOn,
		},
		Refresh: func() (interface{}, string, error) {
			info, err := v.Info().Get(ctx, id)
			if err != nil {
				return "", "", err
			}
			return info, info.Status, nil
		},
	}
	if _, err = vmState.WaitForStateContext(ctx); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
