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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"go.anx.io/go-anxcloud/pkg/client"
	"go.anx.io/go-anxcloud/pkg/vsphere"
	"go.anx.io/go-anxcloud/pkg/vsphere/provisioning/templates"
	"go.anx.io/go-anxcloud/pkg/vsphere/provisioning/vm"
)

// This versioning scheme that currently seems to be in place for template build numbers.
var buildNumberRegex = regexp.MustCompile(`[bB]?(\d+)`)

const (
	templateName = "Flatcar Linux Stable"
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

	templateID, diag := templateIDFromResourceData(
		context.TODO(),
		vsphere.NewAPI(integrationTestClientFromEnv(t)),
		schema.TestResourceDataRaw(t, schemaVirtualServer(), map[string]interface{}{
			"template":    templateName,
			"location_id": envInfo.Location,
		}),
	)
	if diag.HasError() {
		t.Fatalf("failed to retrieve template: %#v\n", diag)
	}

	vmDef := vm.Definition{
		Location:           envInfo.Location,
		Hostname:           fmt.Sprintf("terraform-test-%s-create-virtual-server", envInfo.TestRunName),
		TemplateID:         templateID,
		TemplateType:       "templates",
		Memory:             2048,
		CPUs:               2,
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
	vmDefUpscale.CPUs = 4
	vmDefUpscale.Memory = 4096

	// down scale resources which does not require recreation of the VM
	vmDefDownscale := vmDefUpscale
	vmDefUpscale.CPUs = 2
	vmDefDownscale.Memory = 3072

	vmAddTag := vmDef

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAnxCloudVirtualServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigAnxCloudVirtualServer(resourceName, templateName, &vmDef),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnxCloudVirtualServerExists(resourcePath, &vmDef),
					resource.TestCheckResourceAttr(resourcePath, "location_id", vmDef.Location),
					resource.TestCheckResourceAttr(resourcePath, "template_id", vmDef.TemplateID),
					resource.TestCheckResourceAttr(resourcePath, "cpus", strconv.Itoa(vmDef.CPUs)),
					resource.TestCheckResourceAttr(resourcePath, "memory", strconv.Itoa(vmDef.Memory)),
				),
			},
			{
				Config: testAccConfigAnxCloudVirtualServer(resourceName, templateName, &vmAddTag, "newTag"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnxCloudVirtualServerExists(resourcePath, &vmAddTag),
					resource.TestCheckResourceAttr(resourcePath, "tags.0", "newTag"),
				),
			},
			{
				Config: testAccConfigAnxCloudVirtualServer(resourceName, templateName, &vmDefUpscale),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnxCloudVirtualServerExists(resourcePath, &vmDefUpscale),
					resource.TestCheckResourceAttr(resourcePath, "location_id", vmDefUpscale.Location),
					resource.TestCheckResourceAttr(resourcePath, "template_id", vmDefUpscale.TemplateID),
					resource.TestCheckResourceAttr(resourcePath, "cpus", strconv.Itoa(vmDefUpscale.CPUs)),
					resource.TestCheckResourceAttr(resourcePath, "memory", strconv.Itoa(vmDefUpscale.Memory)),
				),
			},
			{
				Config: testAccConfigAnxCloudVirtualServer(resourceName, templateName, &vmDefDownscale),
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
				ImportStateVerifyIgnore: []string{"critical_operation_confirmed", "enter_bios_setup", "force_restart_if_needed", "hostname", "password", "template", "template_type", "network"},
			},
			{
				Config: testAccConfigAnxCloudVirtualServer(resourceName, templateName, &vmAddTag, "newTag"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnxCloudVirtualServerExists(resourcePath, &vmAddTag),
					resource.TestCheckResourceAttr(resourcePath, "tags.0", "newTag"),
				),
			},
			{
				Config: testAccConfigAnxCloudVirtualServer(resourceName, templateName, &vmDefUpscale),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnxCloudVirtualServerExists(resourcePath, &vmDefUpscale),
					resource.TestCheckResourceAttr(resourcePath, "location_id", vmDefUpscale.Location),
					resource.TestCheckResourceAttr(resourcePath, "template_id", vmDefUpscale.TemplateID),
					resource.TestCheckResourceAttr(resourcePath, "cpus", strconv.Itoa(vmDefUpscale.CPUs)),
					resource.TestCheckResourceAttr(resourcePath, "memory", strconv.Itoa(vmDefUpscale.Memory)),
				),
			},
			{
				Config: testAccConfigAnxCloudVirtualServer(resourceName, templateName, &vmDefDownscale),
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
				ImportStateVerifyIgnore: []string{"critical_operation_confirmed", "enter_bios_setup", "force_restart_if_needed", "hostname", "password", "template", "template_type", "network"},
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
	templateID := vsphereAccTestTemplateByLocationAndPrefix(envInfo.Location, templateName)
	vmDef := vm.Definition{
		Location:           envInfo.Location,
		TemplateType:       "templates",
		TemplateID:         templateID,
		Hostname:           fmt.Sprintf("terraform-test-%s-multi-disk-scaling", envInfo.TestRunName),
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
		addDiskDef.Hostname = fmt.Sprintf("terraform-test-%s-add-disk", envInfo.TestRunName)
		addDiskDef.Network = []vm.Network{createNewNetworkInterface(envInfo)}

		disksAdd := append(disks, vm.Disk{

			Type:    "ENT6",
			SizeGBs: 50,
		})

		resource.ParallelTest(t, resource.TestCase{
			PreCheck:          func() { testAccPreCheck(t) },
			ProviderFactories: testAccProviderFactories,
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
		changeDiskDef.Hostname = fmt.Sprintf("terraform-test-%s-change-add-disk", envInfo.TestRunName)
		changeDiskDef.Network = []vm.Network{createNewNetworkInterface(envInfo)}
		vmRecorder.RecordVMByName(fmt.Sprintf("%%-%s", changeDiskDef.Hostname))
		disksChange := append(disks, vm.Disk{
			SizeGBs: 50,
			Type:    "ENT6",
		})

		disksChange[0].SizeGBs = 70
		disksChange[0].Type = "ENT1"

		resource.ParallelTest(t, resource.TestCase{
			PreCheck:          func() { testAccPreCheck(t) },
			ProviderFactories: testAccProviderFactories,
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
		changeDiskDef.Hostname = fmt.Sprintf("terraform-test-%s-multi-disk-template-change", envInfo.TestRunName)
		changeDiskDef.Network = []vm.Network{createNewNetworkInterface(envInfo)}
		vmRecorder.RecordVMByName(fmt.Sprintf("%%-%s", changeDiskDef.Hostname))
		changeDiskDef.TemplateID = vsphereAccTestTemplateByLocationAndPrefix(envInfo.Location, "Flatcar Storage Stable")
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

		resource.ParallelTest(t, resource.TestCase{
			PreCheck:          func() { testAccPreCheck(t) },
			ProviderFactories: testAccProviderFactories,
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
	c := testAccProvider.Meta().(providerContext).legacyClient
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
func testAccConfigAnxCloudVirtualServer(resourceName string, templateName string, def *vm.Definition, tags ...string) string {
	return fmt.Sprintf(`
	resource "anxcloud_virtual_server" "%s" {
		location_id          = "%s"
		template             = "%s"
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
	`, resourceName, def.Location, templateName, def.Hostname, def.CPUs, def.CPUPerformanceType, def.Memory,
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
		c := testAccProvider.Meta().(providerContext).legacyClient
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
		if info.CPUPerformanceType != def.CPUPerformanceType {
			return fmt.Errorf("virtual machine cpu_performance_type does not match, got %s - expected %s", info.CPUPerformanceType, def.CPUPerformanceType)
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
		c := testAccProvider.Meta().(providerContext).legacyClient
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

func vsphereAccTestTemplateByLocationAndPrefix(locationID string, templateNamePrefix string) string {
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
		if strings.HasPrefix(tpl.Name, templateNamePrefix) {
			selected = append(selected, tpl)
		}
	}

	if len(selected) < 1 {
		log.Fatalf("Template with prefix '%s' not found at location with ID '%s'", templateNamePrefix, locationID)
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

func mockedTemplateList() []templates.Template {
	return []templates.Template{
		{ID: "e9325be9-25b9-468e-851e-56b5c0367e5a", Name: "Ubuntu 21.04", Build: "b72"},
		{ID: "b21b8b77-30e3-478a-9b6d-1f61d29e9f9a", Name: "Flatcar Linux Stable", Build: "b73"},
		{ID: "ec547552-d453-42e6-987d-51abe703c439", Name: "Debian 11", Build: "b18"},
		{ID: "26a47eee-dc9a-4eea-b67a-8fb1baa2fcc0", Name: "Flatcar Linux Stable", Build: "b74"},
		{ID: "cb16dc94-ec55-4e9a-a1a3-b76a91bbe274", Name: "Windows 2022", Build: "b06"},
		{ID: "fc3a63c6-6f4e-4193-b368-ebe9e08b4302", Name: "Debian 10", Build: "b80"},
		{ID: "844ac596-5f62-4ed2-936e-b99ffe0d4f88", Name: "Flatcar Linux Stable", Build: "b72"},
		{ID: "c3d4f0a6-978a-49fb-a952-7361bf531e4f", Name: "Debian 9", Build: "b92"},
		{ID: "086c5f99-1be6-46ec-8374-cdc23cedd6a4", Name: "Windows 2022", Build: "b12"},
		{ID: "9d863fd9-d0d3-4959-b226-e73192f3e43d", Name: "Debian 11", Build: "possibly-valid-build-id"},
	}
}

func TestFindNamedTemplate(t *testing.T) {
	type testCase struct {
		expectedID         string
		expectExisting     bool
		namedTemplate      string
		namedTemplateBuild string
	}

	testCases := []testCase{
		// valid test cases
		{"844ac596-5f62-4ed2-936e-b99ffe0d4f88", true, "Flatcar Linux Stable", "b72"},
		{"26a47eee-dc9a-4eea-b67a-8fb1baa2fcc0", true, "Flatcar Linux Stable", "latest"},
		{"26a47eee-dc9a-4eea-b67a-8fb1baa2fcc0", true, "Flatcar Linux Stable", ""},
		{"26a47eee-dc9a-4eea-b67a-8fb1baa2fcc0", true, "Flatcar Linux Stable", "b74"},
		{"b21b8b77-30e3-478a-9b6d-1f61d29e9f9a", true, "Flatcar Linux Stable", "b73"},
		{"086c5f99-1be6-46ec-8374-cdc23cedd6a4", true, "Windows 2022", "latest"},
		{"086c5f99-1be6-46ec-8374-cdc23cedd6a4", true, "Windows 2022", "b12"},
		{"cb16dc94-ec55-4e9a-a1a3-b76a91bbe274", true, "Windows 2022", "b06"},
		{"9d863fd9-d0d3-4959-b226-e73192f3e43d", true, "Debian 11", "possibly-valid-build-id"},

		// non-existing template name
		{"", false, "FooOS 22.05", "b01"},
		{"", false, "FooOS 22.05", "b06"},
		{"", false, "Bar OS 95", "latest"},

		// non-existing build id
		{"", false, "Windows 2022", "foo"},
		{"", false, "Windows 2022", "b00"},
	}

	for _, testCase := range testCases {
		if id, diag := findNamedTemplate(testCase.namedTemplate, testCase.namedTemplateBuild, mockedTemplateList()); testCase.expectExisting == (diag != nil) {
			t.Errorf("unexpected error: %v", diag)
		} else if id != testCase.expectedID {
			t.Errorf("identifier %q expected, got %q", testCase.expectedID, id)
		}
	}

}
