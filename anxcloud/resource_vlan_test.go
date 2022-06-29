package anxcloud

import (
	"context"
	"fmt"
	"testing"

	"github.com/anexia-it/terraform-provider-anxcloud/anxcloud/testutils/environment"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"go.anx.io/go-anxcloud/pkg/vlan"
)

func TestAccAnxCloudVLAN(t *testing.T) {
	environment.SkipIfNoEnvironment(t)

	resourceName := "acc_test"
	resourcePath := "anxcloud_vlan." + resourceName

	locationID := environment.GetEnvInfo(t).Location
	customerDescription := "vlan acceptance tests"
	customerDescriptionUpdate := "vlan acceptance tests update " + environment.GetEnvInfo(t).TestRunName

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAnxCloudVLANDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAnxCloudVLAN(resourceName, locationID, customerDescription, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "location_id", locationID),
					resource.TestCheckResourceAttr(resourcePath, "description_customer", customerDescription),
					testAccCheckAnxCloudVLANExists(resourcePath, customerDescription, true),
				),
			},
			{
				Config: testAccCheckAnxCloudVLAN(resourceName, locationID, customerDescriptionUpdate, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "location_id", locationID),
					resource.TestCheckResourceAttr(resourcePath, "description_customer", customerDescriptionUpdate),
					testAccCheckAnxCloudVLANExists(resourcePath, customerDescriptionUpdate, false),
				),
			},
			{
				Config: testAccCheckAnxCloudVLAN(resourceName, locationID, customerDescriptionUpdate, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "location_id", locationID),
					resource.TestCheckResourceAttr(resourcePath, "description_customer", customerDescriptionUpdate),
					testAccCheckAnxCloudVLANExists(resourcePath, customerDescriptionUpdate, true),
				),
			},
			{
				ResourceName:      resourcePath,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAnxCloudVLANTags(t *testing.T) {
	environment.SkipIfNoEnvironment(t)
	envInfo := environment.GetEnvInfo(t)

	tpl := fmt.Sprintf(`
	resource "anxcloud_vlan" "foo" {
		location_id = "%s"
		description_customer = "tf-acc-tags"

		%%s // tags
	}`, envInfo.Location)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAnxCloudVLANDestroy,
		Steps: testAccAnxCloudCommonResourceTagTestSteps(
			tpl, "anxcloud_vlan.foo",
		),
	})
}

func testAccCheckAnxCloudVLANDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(providerContext).legacyClient
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
		if info.Status != vlanStatusDeleted {
			return fmt.Errorf("vlan '%s' exists", info.Identifier)
		}
	}

	return nil
}

func testAccCheckAnxCloudVLAN(resourceName, locationID, customerDescription string, vmProvisioning bool) string {
	return fmt.Sprintf(`
	resource "anxcloud_vlan" "%s" {
		location_id   = "%s"
		vm_provisioning = %t
		description_customer = "%s"
	}
	`, resourceName, locationID, vmProvisioning, customerDescription)
}

func testAccCheckAnxCloudVLANExists(n string, expectedCustomerDescription string, expectedVMProvisioning bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		c := testAccProvider.Meta().(providerContext).legacyClient
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
		if i.VMProvisioning != expectedVMProvisioning {
			return fmt.Errorf("vm_provisioning is different than expected '%t': '%t'", i.VMProvisioning, expectedVMProvisioning)
		}

		return nil
	}
}
