package anxcloud

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/anexia-it/go-anxcloud/pkg/client"
	"github.com/anexia-it/go-anxcloud/pkg/vsphere"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/lithammer/shortuuid"
)

func TestAccAnxcloudVirtualServerBasic(t *testing.T) {
	resourceName := "acc_test_basic"
	resourcePath := "anxcloud_virtual_server." + resourceName

	locationID := "52b5f6b2fd3a4a7eaaedf1a7c019e9ea"
	templateID := "12c28aa7-604d-47e9-83fb-5f1d1f1837b3"
	vlanID := "02f39d20ca0f4adfb5032f88dbc26c39"
	cpus := 4
	memory := 4096
	diskSize := 50

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckAnxcloudVirtualServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAnxcloudVirtualServerBasic(resourceName, locationID, templateID, vlanID, cpus, memory, diskSize),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnxcloudVirtualServerExists(resourcePath),
					resource.TestCheckResourceAttr(resourcePath, "location_id", locationID),
					resource.TestCheckResourceAttr(resourcePath, "template_id", templateID),
					resource.TestCheckResourceAttr(resourcePath, "cpus", strconv.Itoa(cpus)),
					resource.TestCheckResourceAttr(resourcePath, "memory", strconv.Itoa(memory)),
					resource.TestCheckResourceAttr(resourcePath, "disk", strconv.Itoa(diskSize)),
				),
			},
		},
	})
}

func testAccCheckAnxcloudVirtualServerDestroy(s *terraform.State) error {
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

func testAccCheckAnxcloudVirtualServerBasic(resourceName, locationID, templateID, vlanID string, cpus, memory, diskSize int) string {
	uuid := shortuuid.New()
	return fmt.Sprintf(`
	resource "anxcloud_virtual_server" "%s" {
		location_id   = "%s"
		template_id   = "%s"
		template_type = "templates"
		hostname      = "acc-test-basic-%s"
		cpus          = %d
		memory        = %d
		disk          = %d
		password      = "flatcar#1234$%%"

		network {
			vlan_id  = "%s"
			nic_type = "vmxnet3"
		}
	}
	`, resourceName, locationID, templateID, uuid, cpus, memory, diskSize, vlanID)
}

func testAccCheckAnxcloudVirtualServerExists(n string) resource.TestCheckFunc {
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

		return nil
	}
}
