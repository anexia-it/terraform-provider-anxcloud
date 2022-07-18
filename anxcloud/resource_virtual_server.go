package anxcloud

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"go.anx.io/go-anxcloud/pkg/vsphere/provisioning/progress"
	"go.anx.io/go-anxcloud/pkg/vsphere/provisioning/templates"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"go.anx.io/go-anxcloud/pkg/ipam/address"
	"go.anx.io/go-anxcloud/pkg/vsphere"
	"go.anx.io/go-anxcloud/pkg/vsphere/provisioning/nictype"
	"go.anx.io/go-anxcloud/pkg/vsphere/provisioning/vm"
)

const (
	maxDNSEntries = 4
	vmPoweredOn   = "poweredOn"
	vmPoweredOff  = "poweredOff"
)

func resourceVirtualServer() *schema.Resource {
	return &schema.Resource{
		Description: `
The virtual_server resource allows you to configure and run virtual machines.

### Known limitations
- removal of disks not supported
- removal of networks not supported
`,
		CreateContext: tagsMiddlewareCreate(resourceVirtualServerCreate),
		ReadContext:   tagsMiddlewareRead(resourceVirtualServerRead),
		UpdateContext: tagsMiddlewareUpdate(resourceVirtualServerUpdate),
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
		Schema: withTagsAttribute(schemaVirtualServer()),
		CustomizeDiff: customdiff.All(
			customdiff.ForceNewIf("template_id", func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) bool {
				// prevent ForceNew when vm-template is controlled by (named) "template" parameter
				_, exist := d.GetOkExists("template")
				return !exist
			}),
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

	provContext := m.(providerContext)
	vsphereAPI := vsphere.NewAPI(provContext.legacyClient)
	addressAPI := address.NewAPI(provContext.legacyClient)
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

	templateID, _diags := templateIDFromResourceData(ctx, vsphereAPI, d)
	diags = append(diags, _diags...)

	templateType := "templates"
	if _, isNamedTemplate := d.GetOk("template"); !isNamedTemplate {
		templateType = d.Get("template_type").(string)
	}

	if len(diags) > 0 {
		return diags
	}

	def := vm.Definition{
		Location:           locationID,
		TemplateType:       templateType,
		TemplateID:         templateID,
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
	provisioning, err := vsphereAPI.Provisioning().VM().Provision(ctx, def, base64Encoding)
	if err != nil {
		return diag.FromErr(err)
	}

	vmIdentifier, err := vsphereAPI.Provisioning().Progress().AwaitCompletion(ctx, provisioning.Identifier)
	if err != nil {
		return diag.Errorf("failed to await completion: %s", err)
	}

	d.SetId(vmIdentifier)

	if len(disks) > 1 {
		if read := resourceVirtualServerRead(ctx, d, m); read.HasError() {
			return read
		}

		initialDisks := expandVirtualServerDisks(d.Get("disk").([]interface{}))
		if update := updateVirtualServerDisk(ctx, provContext, d.Id(), disks, initialDisks); update != nil {
			return update
		}
	}

	diags = resourceVirtualServerRead(ctx, d, m)
	return diags
}

func resourceVirtualServerRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(providerContext).legacyClient
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

	if err = d.Set("cpu_performance_type", info.CPUPerformanceType); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
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
	provContext := m.(providerContext)
	vsphereAPI := vsphere.NewAPI(provContext.legacyClient)
	ch := vm.Change{
		Reboot:          d.Get("force_restart_if_needed").(bool),
		EnableDangerous: d.Get("critical_operation_confirmed").(bool),
	}

	if d.HasChanges("sockets", "memory", "cpus") {
		ch.CPUs = d.Get("cpus").(int)
		ch.CPUSockets = d.Get("sockets").(int)
		ch.MemoryMBs = d.Get("memory").(int)
	}

	// cpu_performance_type might not be set because info endpoint didn't expose it previously
	// therefore only change it when the argument changes
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

	provisioning, err := vsphereAPI.Provisioning().VM().Update(ctx, d.Id(), ch)
	if err != nil {
		return diag.FromErr(err)
	}

	if _, err = vsphereAPI.Provisioning().Progress().AwaitCompletion(ctx, provisioning.Identifier); err != nil {
		return diag.FromErr(err)
	}

	return resourceVirtualServerRead(ctx, d, m)
}

func resourceVirtualServerDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(providerContext).legacyClient
	vsphereAPI := vsphere.NewAPI(c)
	progressAPI := progress.NewAPI(c)

	delayedDeprovision := false
	response, err := vsphereAPI.Provisioning().VM().Deprovision(ctx, d.Id(), delayedDeprovision)
	if err != nil {
		if err := handleNotFoundError(err); err != nil {
			return diag.FromErr(err)
		}
		d.SetId("")
		return nil
	}

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		response, err := progressAPI.Get(ctx, response.Identifier)
		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("failed to fetch deprovison progress: %w", err))
		}

		if len(response.Errors) > 0 {
			joinedErrors := strings.Join(response.Errors, ",")
			return resource.NonRetryableError(fmt.Errorf("errors during deprovision: [%s]", joinedErrors))
		}

		if response.Progress == 100 {
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

func updateVirtualServerDisk(ctx context.Context, m providerContext, id string, expected []Disk, current []Disk) diag.Diagnostics {
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
			logger.Error(nil, "Skipping disk %d because expected disk size to small! Expected: %d  -  got: %f", actualDisk.ID, expectedDisk.SizeGBs, actualDisk.ExactDiskSize)
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

	v := vsphere.NewAPI(m.legacyClient)
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

func templateIDFromResourceData(ctx context.Context, a vsphere.API, d *schema.ResourceData) (string, diag.Diagnostics) {
	if templateID, ok := d.GetOk("template_id"); ok {
		return templateID.(string), nil
	}

	// TODO: templates pagination is currently broken (see comments in ENGSUP-4364)
	// template count is far from 1K but this needs proper pagination as soon as ADC API 2.0 is available
	templates, err := a.Provisioning().Templates().List(ctx, d.Get("location_id").(string), "templates", 1, 1000)
	if err != nil {
		return "", diag.FromErr(err)
	}

	return findNamedTemplate(d.Get("template").(string), d.Get("template_build").(string), templates)
}

func findNamedTemplate(name, build string, tpls []templates.Template) (string, diag.Diagnostics) {
	var (
		match   = -1
		buildNo = -1
		latest  = build == "" || build == "latest"
	)

	for i, template := range tpls {
		if template.Name != name {
			continue
		}

		if latest {
			currentTemplateBuildNo, _ := strconv.Atoi(template.Build[1:])

			if latest && (match < 0 || currentTemplateBuildNo > buildNo) {
				match = i
				buildNo = currentTemplateBuildNo
			}
		} else if template.Build == build {
			match = i
			break
		}

	}

	if match < 0 {
		return "", diag.Errorf("named template %q with %q build wasn't found at the specified location", name, build)
	}

	return tpls[match].ID, nil
}
