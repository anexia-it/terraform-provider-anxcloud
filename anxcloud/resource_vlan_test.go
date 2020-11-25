package anxcloud

import (
	"context"
	"fmt"
	"testing"

	"github.com/anexia-it/go-anxcloud/pkg/client"
	"github.com/anexia-it/go-anxcloud/pkg/vlan"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAnxcloudVLAN(t *testing.T) {
	resourceName := "acc_test_basic"
	resourcePath := "anxcloud_vlan." + resourceName

	locationID := "52b5f6b2fd3a4a7eaaedf1a7c019e9ea"
	customerDescription := "vlan acceptance tests"
	customerDescriptionUpdate := "vlan acceptance tests update"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckAnxcloudVLANDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAnxcloudVLAN(resourceName, locationID, customerDescription),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "location_id", locationID),
					resource.TestCheckResourceAttr(resourcePath, "vm_provisioning", "true"),
					resource.TestCheckResourceAttr(resourcePath, "description_customer", customerDescription),
					testAccCheckAnxcloudVLANExists(resourcePath, customerDescription),
				),
			},
			{
				Config: testAccCheckAnxcloudVLAN(resourceName, locationID, customerDescriptionUpdate),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "location_id", locationID),
					resource.TestCheckResourceAttr(resourcePath, "vm_provisioning", "true"),
					resource.TestCheckResourceAttr(resourcePath, "description_customer", customerDescriptionUpdate),
					testAccCheckAnxcloudVLANExists(resourcePath, customerDescriptionUpdate),
				),
			},
		},
	})
}

func testAccCheckAnxcloudVLANDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(client.Client)
	v := vlan.NewAPI(c)
	ctx := context.Background()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "anxcloud_vlan" {
			continue
		}

		if rs.Primary.ID == "" {
			return nil
		}

		info, err := v.Get(ctx, rs.Primary.ID)
		if err != nil {
			if err := handleNotFoundError(err); err != nil {
				return err
			}
			return nil
		}
		if info.Identifier != "" {
			return fmt.Errorf("vlan '%s' exists", info.Identifier)
		}
	}

	return nil
}

func testAccCheckAnxcloudVLAN(resourceName, locationID, customerDescription string) string {
	return fmt.Sprintf(`
	resource "anxcloud_vlan" "%s" {
		location_id   = "%s"
		vm_provisioning = true
		description_customer = "%s"
	}
	`, resourceName, locationID, customerDescription)
}

func testAccCheckAnxcloudVLANExists(n string, expectedCustomerDescription string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		c := testAccProvider.Meta().(client.Client)
		v := vlan.NewAPI(c)
		ctx := context.Background()

		if !ok {
			return fmt.Errorf("vlan not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("vlan id not set")
		}

		i, err := v.Get(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		if i.Status != vlanStatusActive {
			return fmt.Errorf("vlan found but it is not in the expected state '%s': %s", vlanStatusActive, i.Status)
		}

		if i.CustomerDescription != expectedCustomerDescription {
			return fmt.Errorf("customer description is different than expected '%s': '%s'", i.CustomerDescription, expectedCustomerDescription)
		}

		return nil
	}
}
