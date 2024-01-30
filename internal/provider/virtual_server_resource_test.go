//nolint:unparam
package provider

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"text/template"

	"github.com/anexia-it/terraform-provider-anxcloud/anxcloud/testutils/environment"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"go.anx.io/go-anxcloud/pkg/client"
	"go.anx.io/go-anxcloud/pkg/vsphere"
)

type virtualServerResourceData struct {
	Hostname string

	Template     string
	TemplateID   string
	TemplateType string

	Location string

	CPUs               int
	CPUPerformanceType string
	Sockets            int
	Memory             int
	DNS                *[]string
	Script             string

	Disks    []virtualServerResourceDataDisk
	Networks []virtualServerResourceDataNetwork

	ForceRestartIfNeeded       bool
	CriticalOperationConfirmed bool
}

type virtualServerResourceDataDisk struct {
	SizeGB int
	Type   string
}

type virtualServerResourceDataNetwork struct {
	VLAN    string
	IPs     []string
	NICType string
}

func (d virtualServerResourceData) toTerraform(location, templateName string) string {
	var out strings.Builder

	tmpl := template.Must(template.New("virtual server").Parse(`
	data "anxcloud_core_location" "foo" {
	  code = "{{ .Location }}"
	}

	{{ if .Template }}
	data "anxcloud_virtual_server_template" "foo" {
	  name     = "{{ .Template }}"
	  location = data.anxcloud_core_location.foo.id
	}
	{{ end }}

	resource "anxcloud_virtual_server" "foo" {
		hostname  = "{{ .Hostname }}"
		cpus      = {{ .CPUs }}
		cpu_performance_type = "{{ .CPUPerformanceType }}"
		memory    = {{ .Memory }}

		{{ if gt .Sockets 0 }}
		sockets = {{ .Sockets }}
		{{ end }}

		{{ if .TemplateID }}
		template_id = "{{ .TemplateID }}"
		{{ else }}
		template_id = data.anxcloud_virtual_server_template.foo.id
		{{ end }}
		template_type = "{{ .TemplateType }}"


		location_id = data.anxcloud_core_location.foo.id

		{{ range .Disks }}
		disk {
			disk_gb   = {{ .SizeGB }}
			disk_type = "{{ .Type }}"
		}
		{{ end }}

		{{ range .Networks }}
		network {
			vlan_id = "{{ .VLAN }}"
			ips = [
				{{ range .IPs }}
				"{{ . }}",
				{{ end}}
			]
			nic_type = "{{ .NICType }}"
		}
		{{ end }}

		password = "flatcar#1234$%%"

		{{ if .ForceRestartIfNeeded }}
		force_restart_if_needed = true
		{{ end }}

		{{ if .CriticalOperationConfirmed }}
		critical_operation_confirmed = true
		{{ end }}
	}
	`))

	if err := tmpl.Execute(&out, d); err != nil {
		panic(err)
	}

	return out.String()
}

func TestAccVirtualServerResource(t *testing.T) {
	environment.SkipIfNoEnvironment(t)
	envInfo := environment.GetEnvInfo(t)

	config := virtualServerResourceData{
		Hostname:           fmt.Sprintf("terraform-test-%s", envInfo.TestRunName),
		Location:           "ANX04",
		CPUs:               4,
		CPUPerformanceType: "standard",
		Memory:             2048,
		Template:           "Debian 11",
		TemplateType:       "templates",
		Disks: []virtualServerResourceDataDisk{
			{SizeGB: 20, Type: "STD4"},
			{SizeGB: 15, Type: "STD2"},
		},
		Networks: []virtualServerResourceDataNetwork{
			{
				VLAN:    envInfo.VlanID,
				NICType: "vmxnet3",
				IPs: []string{
					envInfo.Prefix.GetNextIP(),
					envInfo.Prefix.GetNextIP(),
				},
			},
		},
	}

	changedConfig := config
	changedConfig.CPUs = 2
	changedConfig.Sockets = 2
	changedConfig.Memory = 4096
	changedConfig.Disks = append(config.Disks, virtualServerResourceDataDisk{
		SizeGB: 30,
		Type:   "ENT2",
	})

	changedConfigAllowCriticalAndRestart := config
	changedConfigAllowCriticalAndRestart.ForceRestartIfNeeded = true
	changedConfigAllowCriticalAndRestart.CriticalOperationConfirmed = true

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config.toTerraform("ANX04", "Debian 11"),
				Check:  testAccCheckVirtualServerResourceExists(t, "anxcloud_virtual_server.foo", config),
			},
			{
				Config:      changedConfig.toTerraform("ANX04", "Debian 11"),
				Check:       testAccCheckVirtualServerResourceExists(t, "anxcloud_virtual_server.foo", changedConfig),
				ExpectError: regexp.MustCompile("VM has to be powered off"),
			},
			{
				Config: changedConfigAllowCriticalAndRestart.toTerraform("ANX04", "Debian 11"),
				Check:  testAccCheckVirtualServerResourceExists(t, "anxcloud_virtual_server.foo", changedConfigAllowCriticalAndRestart),
			},
			{
				ResourceName:      "anxcloud_virtual_server.foo",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"hostname",                     // implements semantic equality (not covered by import state verification)
					"cpu_performance_type",         // implements semantic equality (not covered by import state verification)
					"password",                     // field is not returned by API
					"critical_operation_confirmed", // field is only used for resource updates
					"force_restart_if_needed",      // field is only used for resource updates
				},
			},
		},
	})
}

func TestAccVirtualServerResourceFromScratch(t *testing.T) {
	environment.SkipIfNoEnvironment(t)
	envInfo := environment.GetEnvInfo(t)

	config := virtualServerResourceData{
		Hostname:           fmt.Sprintf("terraform-test-%s-from-scratch", envInfo.TestRunName),
		Location:           "ANX04",
		CPUs:               4,
		CPUPerformanceType: "standard",
		Memory:             2048,
		TemplateID:         "114",
		TemplateType:       "from_scratch",
		Disks: []virtualServerResourceDataDisk{
			{SizeGB: 20, Type: "STD4"},
			{SizeGB: 15, Type: "STD2"},
		},
		Networks: []virtualServerResourceDataNetwork{
			{
				VLAN:    envInfo.VlanID,
				NICType: "vmxnet3",
				IPs: []string{
					envInfo.Prefix.GetNextIP(),
					envInfo.Prefix.GetNextIP(),
				},
			},
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config.toTerraform("ANX04", "Debian 11"),
				Check:  testAccCheckVirtualServerResourceExists(t, "anxcloud_virtual_server.foo", config),
			},
			{
				ResourceName: "anxcloud_virtual_server.foo",
				ImportState:  true,
				ExpectError:  regexp.MustCompile("Cannot import virtual server with `from_scratch` template"),
			},
		},
	})
}

func testClient(t *testing.T) client.Client {
	t.Helper()
	client, err := client.New(client.AuthFromEnv(false))
	if err != nil {
		t.Error(err)
	}

	return client
}

func testAccCheckVirtualServerResourceExists(t *testing.T, n string, config virtualServerResourceData) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("virtual server not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("virtual server id not set")
		}

		vsphereAPI := vsphere.NewAPI(testClient(t))
		info, err := vsphereAPI.Info().Get(context.TODO(), rs.Primary.ID)
		if err != nil {
			return err
		}

		if !strings.HasSuffix(info.Name, config.Hostname) {
			return fmt.Errorf("configured virtual machine hostname is not a suffix of actual hostname, got %s - expected %s", info.Name, config.Hostname)
		}
		if info.CPU != config.CPUs {
			return fmt.Errorf("virtual machine cpu does not match, got %d - expected %d", info.CPU, config.CPUs)
		}
		if info.RAM != config.Memory {
			return fmt.Errorf("virtual machine cpu does not match, got %d - expected %d", info.CPU, config.CPUs)
		}
		if !strings.HasPrefix(info.CPUPerformanceType, config.CPUPerformanceType) {
			return fmt.Errorf("cpu_performance_type does not match")
		}

		if len(info.DiskInfo) != len(config.Disks) {
			return fmt.Errorf("unexpected number of disks, got %d - expected %d", len(info.DiskInfo), len(config.Disks))
		}
		for i := range info.DiskInfo {
			if int(info.DiskInfo[i].DiskGB) != config.Disks[i].SizeGB {
				return fmt.Errorf("unexpected disk size for disk with index %d, got %d - expected %d", i, int(info.DiskInfo[i].DiskGB), config.Disks[i].SizeGB)
			}
			if info.DiskInfo[i].DiskType != config.Disks[i].Type {
				return fmt.Errorf("unexpected disk type for disk with index %d, got %q - expected %q", i, info.DiskInfo[i].DiskType, config.Disks[i].Type)
			}
		}

		if len(info.Network) != len(config.Networks) {
			return fmt.Errorf("unexpected number of networks, got %d - expected %d", len(info.Network), len(config.Networks))
		}
		for i := range info.Network {
			if info.Network[i].VLAN != config.Networks[i].VLAN {
				return fmt.Errorf("unexpected disk size for disk with index %d, got %d - expected %d", i, int(info.DiskInfo[i].DiskGB), config.Disks[i].SizeGB)
			}
			// todo: check ips and nictype
		}

		return nil
	}
}
