package anxcloud

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/anexia-it/terraform-provider-anxcloud/anxcloud/testutils/environment"
	"github.com/anexia-it/terraform-provider-anxcloud/anxcloud/testutils/recorder"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"go.anx.io/go-anxcloud/pkg/client"
	"go.anx.io/go-anxcloud/pkg/vsphere"
	"go.anx.io/go-anxcloud/pkg/vsphere/provisioning/templates"
	"go.anx.io/go-anxcloud/pkg/vsphere/provisioning/vm"
)

// This versioning scheme that currently seems to be in place for template build numbers.
var buildNumberRegex = regexp.MustCompile(`[bB]?(\d+)`)

const (
	templateName = "Ubuntu 20.04.02"
)

func getVMRecorder(t *testing.T) recorder.VMRecoder {
	vmRecorder := recorder.VMRecoder{}
	t.Cleanup(func() {
		vmRecorder.Cleanup(context.TODO())
	})
	return vmRecorder
}

func TestAccAnxCloudVirtualServer(t *testing.T) {
	environment.SkipIfNoEnvironment(t)
	resourceName := "acc_test_vm_test"
	resourcePath := "anxcloud_virtual_server." + resourceName

	vmRecorder := getVMRecorder(t)
	envInfo := environment.GetEnvInfo(t)
	templateID := vsphereAccTestInit(envInfo.Location, templateName)
	vmDef := vm.Definition{
		Location:           envInfo.Location,
		TemplateType:       "templates",
		TemplateID:         templateID,
		Hostname:           "terraform-test-" + envInfo.TestRunName,
		Memory:             2048,
		CPUs:               1,
		CPUPerformanceType: "performance",
		Disk:               50,
		DiskType:           "ENT6",
		Network:            []vm.Network{createNewNetworkInterface(envInfo)},
		DNS1:               "8.8.8.8",
		Password:           "flatcar#1234$%%",
	}
	vmRecorder.RecordVMByName(fmt.Sprintf("%%-%s", vmDef.Hostname))

	// upscale resources
	vmDefUpscale := vmDef
	vmDefUpscale.CPUs = 2
	vmDefUpscale.Memory = 4096

	// down scale resources which does not require recreation of the VM
	vmDefDownscale := vmDefUpscale
	vmDefDownscale.Memory = 2096

	vmAddTag := vmDef

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckAnxCloudVirtualServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigAnxCloudVirtualServer(resourceName, &vmDef),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnxCloudVirtualServerExists(resourcePath, &vmDef),
					resource.TestCheckResourceAttr(resourcePath, "location_id", vmDef.Location),
					resource.TestCheckResourceAttr(resourcePath, "template_id", vmDef.TemplateID),
					resource.TestCheckResourceAttr(resourcePath, "cpus", strconv.Itoa(vmDef.CPUs)),
					resource.TestCheckResourceAttr(resourcePath, "memory", strconv.Itoa(vmDef.Memory)),
				),
			},
			{
				Config: testAccConfigAnxCloudVirtualServer(resourceName, &vmAddTag, "newTag"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnxCloudVirtualServerExists(resourcePath, &vmAddTag),
					resource.TestCheckResourceAttr(resourcePath, "tags.0", "newTag"),
				),
			},
			{
				Config: testAccConfigAnxCloudVirtualServer(resourceName, &vmDefUpscale),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnxCloudVirtualServerExists(resourcePath, &vmDefUpscale),
					resource.TestCheckResourceAttr(resourcePath, "location_id", vmDefUpscale.Location),
					resource.TestCheckResourceAttr(resourcePath, "template_id", vmDefUpscale.TemplateID),
					resource.TestCheckResourceAttr(resourcePath, "cpus", strconv.Itoa(vmDefUpscale.CPUs)),
					resource.TestCheckResourceAttr(resourcePath, "memory", strconv.Itoa(vmDefUpscale.Memory)),
				),
			},
			{
				Config: testAccConfigAnxCloudVirtualServer(resourceName, &vmDefDownscale),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnxCloudVirtualServerExists(resourcePath, &vmDefDownscale),
					resource.TestCheckResourceAttr(resourcePath, "location_id", vmDefDownscale.Location),
					resource.TestCheckResourceAttr(resourcePath, "template_id", vmDefDownscale.TemplateID),
					resource.TestCheckResourceAttr(resourcePath, "cpus", strconv.Itoa(vmDefDownscale.CPUs)),
					resource.TestCheckResourceAttr(resourcePath, "memory", strconv.Itoa(vmDefDownscale.Memory)),
				),
			},
			{
				ResourceName:            resourcePath,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"cpu_performance_type", "tags.#", "tags.0", "critical_operation_confirmed", "enter_bios_setup", "force_restart_if_needed", "hostname", "password", "template_type", "network"},
			},
			{
				Config: testAccConfigAnxCloudVirtualServer(resourceName, &vmAddTag, "newTag"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnxCloudVirtualServerExists(resourcePath, &vmAddTag),
					resource.TestCheckResourceAttr(resourcePath, "tags.0", "newTag"),
				),
			},
			{
				Config: testAccConfigAnxCloudVirtualServer(resourceName, &vmDefUpscale),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnxCloudVirtualServerExists(resourcePath, &vmDefUpscale),
					resource.TestCheckResourceAttr(resourcePath, "location_id", vmDefUpscale.Location),
					resource.TestCheckResourceAttr(resourcePath, "template_id", vmDefUpscale.TemplateID),
					resource.TestCheckResourceAttr(resourcePath, "cpus", strconv.Itoa(vmDefUpscale.CPUs)),
					resource.TestCheckResourceAttr(resourcePath, "memory", strconv.Itoa(vmDefUpscale.Memory)),
				),
			},
			{
				Config: testAccConfigAnxCloudVirtualServer(resourceName, &vmDefDownscale),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnxCloudVirtualServerExists(resourcePath, &vmDefDownscale),
					resource.TestCheckResourceAttr(resourcePath, "location_id", vmDefDownscale.Location),
					resource.TestCheckResourceAttr(resourcePath, "template_id", vmDefDownscale.TemplateID),
					resource.TestCheckResourceAttr(resourcePath, "cpus", strconv.Itoa(vmDefDownscale.CPUs)),
					resource.TestCheckResourceAttr(resourcePath, "memory", strconv.Itoa(vmDefDownscale.Memory)),
				),
			},
			{
				ResourceName:            resourcePath,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"cpu_performance_type", "tags.#", "tags.0", "critical_operation_confirmed", "enter_bios_setup", "force_restart_if_needed", "hostname", "password", "template_type", "network"},
			},
		},
	})
}

func TestAccAnxCloudVirtualServerMultiDiskScaling(t *testing.T) {
	environment.SkipIfNoEnvironment(t)
	resourceName := "acc_test_vm_test_multi_disk"
	resourcePath := "anxcloud_virtual_server." + resourceName

	vmRecorder := getVMRecorder(t)
	envInfo := environment.GetEnvInfo(t)
	templateID := vsphereAccTestInit(envInfo.Location, templateName)
	vmDef := vm.Definition{
		Location:           envInfo.Location,
		TemplateType:       "templates",
		TemplateID:         templateID,
		Hostname:           "terraform-test-" + envInfo.TestRunName,
		Memory:             2048,
		CPUs:               2,
		CPUPerformanceType: "performance",
		Network:            []vm.Network{createNewNetworkInterface(envInfo)},
		DNS1:               "8.8.8.8",
		Password:           "flatcar#1234$%%",
	}
	vmRecorder.RecordVMByName(fmt.Sprintf("%%-%s", vmDef.Hostname))

	disks := []vm.Disk{
		{
			Type:    "ENT1",
			SizeGBs: 40,
		},
	}

	t.Run("AddDisk", func(t *testing.T) {
		addDiskDef := vmDef
		addDiskDef.Hostname = "terraform-test-" + envInfo.TestRunName
		addDiskDef.Network = []vm.Network{createNewNetworkInterface(envInfo)}

		disksAdd := append(disks, vm.Disk{

			Type:    "ENT6",
			SizeGBs: 50,
		})

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { testAccPreCheck(t) },
			ProviderFactories: testAccProviders,
			CheckDestroy:      testAccCheckAnxCloudVirtualServerDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccConfigAnxCloudVirtualServerMultiDiskSupport(resourceName, &addDiskDef, disks),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAnxCloudVirtualServerDisks(resourcePath, disks),
					),
				},
				{
					Config: testAccConfigAnxCloudVirtualServerMultiDiskSupport(resourceName, &addDiskDef, disksAdd),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAnxCloudVirtualServerDisks(resourcePath, disksAdd),
					),
				},
			},
		})
	})

	t.Run("ChangeAddDisk", func(t *testing.T) {
		changeDiskDef := vmDef
		changeDiskDef.Hostname = "terraform-test-" + envInfo.TestRunName
		changeDiskDef.Network = []vm.Network{createNewNetworkInterface(envInfo)}
		vmRecorder.RecordVMByName(fmt.Sprintf("%%-%s", changeDiskDef.Hostname))
		disksChange := append(disks, vm.Disk{
			SizeGBs: 50,
			Type:    "ENT6",
		})

		disksChange[0].SizeGBs = 70
		disksChange[0].Type = "ENT1"

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { testAccPreCheck(t) },
			ProviderFactories: testAccProviders,
			CheckDestroy:      testAccCheckAnxCloudVirtualServerDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccConfigAnxCloudVirtualServerMultiDiskSupport(resourceName, &changeDiskDef, disks),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAnxCloudVirtualServerDisks(resourcePath, disks),
					),
				},
				{
					Config: testAccConfigAnxCloudVirtualServerMultiDiskSupport(resourceName, &changeDiskDef, disksChange),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAnxCloudVirtualServerDisks(resourcePath, disksChange),
					),
				},
			},
		})
	})

	t.Run("MultiDiskTemplateChange", func(t *testing.T) {
		changeDiskDef := vmDef
		changeDiskDef.Hostname = "terraform-test-" + envInfo.TestRunName
		changeDiskDef.Network = []vm.Network{createNewNetworkInterface(envInfo)}
		vmRecorder.RecordVMByName(fmt.Sprintf("%%-%s", changeDiskDef.Hostname))
		changeDiskDef.TemplateID = "659b35b5-0060-44de-9f9e-a069ec5f1bca"
		templateDisks := []vm.Disk{
			{
				Type:    "ENT6",
				SizeGBs: 50,
			},
			{
				Type:    "ENT6",
				SizeGBs: 50,
			},
		}

		templateDisksChanged := append(templateDisks, vm.Disk{
			SizeGBs: 70,
			Type:    "ENT1",
		})
		templateDisksChanged[1].SizeGBs = 60

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { testAccPreCheck(t) },
			ProviderFactories: testAccProviders,
			CheckDestroy:      testAccCheckAnxCloudVirtualServerDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccConfigAnxCloudVirtualServerMultiDiskSupport(resourceName, &changeDiskDef, templateDisks),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAnxCloudVirtualServerDisks(resourcePath, templateDisks),
					),
				},
				{
					Config: testAccConfigAnxCloudVirtualServerMultiDiskSupport(resourceName, &changeDiskDef, templateDisksChanged),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAnxCloudVirtualServerDisks(resourcePath, templateDisksChanged),
					),
				},
			},
		})
	})
}

func testAccCheckAnxCloudVirtualServerDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(client.Client)
	v := vsphere.NewAPI(c)
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "anxcloud_virtual_server" {
			continue
		}

		if rs.Primary.ID == "" {
			return nil
		}

		info, err := v.Info().Get(ctx, rs.Primary.ID)
		if err != nil {
			if err := handleNotFoundError(err); err != nil {
				return err
			}
			return nil
		}
		if info.Identifier != "" {
			return fmt.Errorf("virtual machine '%s' exists", info.Identifier)
		}
	}

	return nil
}

//nolint:unparam
func testAccConfigAnxCloudVirtualServer(resourceName string, def *vm.Definition, tags ...string) string {
	return fmt.Sprintf(`
	resource "anxcloud_virtual_server" "%s" {
		location_id          = "%s"
		template_id          = "%s"
		template_type        = "%s"
		hostname             = "%s"
		cpus                 = %d
		cpu_performance_type = "%s"
		memory               = %d
		password             = "%s"

		// generated network string
		%s

		// generated disk string
		%s

		// generated tags
		%s

		force_restart_if_needed = true
		critical_operation_confirmed = true
	}
	`, resourceName, def.Location, def.TemplateID, def.TemplateType, def.Hostname, def.CPUs, def.CPUPerformanceType, def.Memory,
		def.Password, generateNetworkSubResourceString(def.Network), generateDisksSubResourceString([]vm.Disk{
			{
				SizeGBs: def.Disk,
				Type:    def.DiskType,
			},
		}), generateTagsString(tags...))
}

func testAccConfigAnxCloudVirtualServerMultiDiskSupport(resourceName string, def *vm.Definition, disks []vm.Disk) string {
	return fmt.Sprintf(`
	resource "anxcloud_virtual_server" "%s" {
		location_id   = "%s"
		template_id   = "%s"
		template_type = "%s"
		hostname      = "%s"
		cpus          = %d
		memory        = %d
		password      = "%s"

		// generated network string
		%s

		// generated disks string
		%s

		force_restart_if_needed = true
		critical_operation_confirmed = true
	}
	`, resourceName, def.Location, def.TemplateID, def.TemplateType, def.Hostname, def.CPUs, def.Memory,
		def.Password, generateNetworkSubResourceString(def.Network), generateDisksSubResourceString(disks))
}

func testAccCheckAnxCloudVirtualServerExists(n string, def *vm.Definition) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		c := testAccProvider.Meta().(client.Client)
		v := vsphere.NewAPI(c)
		ctx := context.Background()

		if !ok {
			return fmt.Errorf("virtual server not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("virtual server id not set")
		}

		info, err := v.Info().Get(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		if info.Status != vmPoweredOn {
			return fmt.Errorf("virtual machine found but it is not in the expected state '%s': '%s'", vmPoweredOn, info.Status)
		}

		if info.CPU != def.CPUs {
			return fmt.Errorf("virtual machine cpu does not match, got %d - expected %d", info.CPU, def.CPUs)
		}
		if info.RAM != def.Memory {
			return fmt.Errorf("virtual machine memory does not match, got %d - expected %d", info.RAM, def.Memory)
		}

		if len(info.DiskInfo) != 1 {
			return fmt.Errorf("unspported number of attached disks, got %d - expected 1", len(info.DiskInfo))
		}
		infoDiskGB := roundDiskSize(info.DiskInfo[0].DiskGB)
		if infoDiskGB != def.Disk {
			return fmt.Errorf("virtual machine disk size does not match, got %d - expected %d", infoDiskGB, def.Disk)
		}

		if len(info.Network) != len(def.Network) {
			return fmt.Errorf("virtual machine networks number do not match, got %d - expected %d", len(info.Network), len(info.Network))
		}
		for i, n := range def.Network {
			if n.VLAN != info.Network[i].VLAN {
				return fmt.Errorf("virtual machine network[%d].vlan do not match, got %s - expected %s", i, info.Network[i].VLAN, n.VLAN)
			}
		}

		return nil
	}
}

func testAccCheckAnxCloudVirtualServerDisks(n string, expectedDisks []vm.Disk) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		c := testAccProvider.Meta().(client.Client)
		v := vsphere.NewAPI(c)
		ctx := context.Background()

		if !ok {
			return fmt.Errorf("virtual server not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("virtual server id not set")
		}

		info, err := v.Info().Get(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		if len(info.DiskInfo) != len(expectedDisks) {
			return fmt.Errorf("virtual machine disk count do not match, got %d - expected %d", len(info.DiskInfo), len(expectedDisks))
		}

		for i, disk := range info.DiskInfo {
			if disk.DiskType != expectedDisks[i].Type {
				return fmt.Errorf("virtual machine disk with ID %d has incorrect type, got %s - expected %s", disk.DiskID, disk.DiskType, expectedDisks[i].Type)
			} else if roundDiskSize(disk.DiskGB) != expectedDisks[i].SizeGBs {
				return fmt.Errorf("virtual machine disk with ID %d has incorrect size, got %f - expected %d", disk.DiskID, disk.DiskGB, expectedDisks[i].SizeGBs)
			}
		}

		return nil
	}
}

func generateNetworkSubResourceString(networks []vm.Network) string {
	var output string
	template := "\nnetwork {\n\tvlan_id = \"%s\"\n\tnic_type = \"%s\"\n\tips = [\"%s\"]\n}\n"

	for _, n := range networks {
		output += fmt.Sprintf(template, n.VLAN, n.NICType, n.IPs[0])
	}

	return output
}

func generateDisksSubResourceString(disks []vm.Disk) string {
	var output string
	template := "\ndisk {\n\tdisk_gb = %d\n\tdisk_type = \"%s\"\n}\n"

	for _, d := range disks {
		output += fmt.Sprintf(template, d.SizeGBs, d.Type)
	}

	return output
}

func generateTagsString(tags ...string) string {
	if len(tags) == 0 {
		return ""
	}

	for i, tag := range tags {
		tags[i] = fmt.Sprintf("\"%s\",", tag)
	}
	return fmt.Sprintf("tags = [\n%s\n]", strings.Join(tags, "\n"))
}

func vsphereAccTestInit(locationID string, templateName string) string {
	if _, ok := os.LookupEnv(client.TokenEnvName); !ok {
		// we are running in unit test environment so do nothing
		return ""
	}
	cli, err := client.New(client.AuthFromEnv(false))
	if err != nil {
		log.Fatalf("Error creating client for retrieving template ID: %v\n", err)
	}

	tplAPI := templates.NewAPI(cli)
	tpls, err := tplAPI.List(context.TODO(), locationID, templates.TemplateTypeTemplates, 1, 500)

	if err != nil {
		log.Fatalf("Error retrieving templates: %v\n", err)
	}

	selected := make([]templates.Template, 0, 1)
	for _, tpl := range tpls {
		if strings.HasPrefix(tpl.Name, templateName) {
			selected = append(selected, tpl)
		}
	}

	sort.Slice(selected, func(i, j int) bool {
		return extractBuildNumber(selected[i].Build) > extractBuildNumber(selected[j].Build)
	})

	log.Printf("VSphere: selected template %v (build %v, ID %v)\n", selected[0].Name, selected[0].Build, selected[0].ID)

	return selected[0].ID
}

func extractBuildNumber(version string) int {
	match := buildNumberRegex.FindStringSubmatch(version)
	if len(match) != 2 {
		panic("the version doesn't match the given regex")
	}
	number, err := strconv.ParseInt(match[1], 10, 0)
	if err != nil {
		panic(fmt.Sprintf("could not extract version for %s", version))
	}
	return int(number)
}

func TestVersionParsing(t *testing.T) {
	require.Equal(t, 5555, extractBuildNumber("b5555"))
	require.Equal(t, 6666, extractBuildNumber("6666"))
}

func createNewNetworkInterface(info environment.Info) vm.Network {
	return vm.Network{
		VLAN:    info.VlanID,
		NICType: "vmxnet3",
		IPs:     []string{info.Prefix.GetNextIP()},
	}
}
