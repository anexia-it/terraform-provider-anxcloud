package anxcloud

import (
	"context"
	"fmt"
	"testing"

	"github.com/anexia-it/terraform-provider-anxcloud/anxcloud/testutils/environment"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"go.anx.io/go-anxcloud/pkg/ipam/address"
)

func TestAccAnxCloudIPAddress(t *testing.T) {
	environment.SkipIfNoEnvironment(t)
	resourceName := "acc_test"
	resourcePath := "anxcloud_ip_address." + resourceName

	envInfo := environment.GetEnvInfo(t)

	prefixID := envInfo.Prefix.ID
	ipAddress := envInfo.Prefix.GetNextIP()
	role := "Default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccAnxCloudIPAddressDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAnxCloudIPAddress(resourceName, prefixID, ipAddress, role),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "network_prefix_id", prefixID),
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

func TestAccAnxCloudIPAddressReserved(t *testing.T) {
	resourceName := "acc_test"
	resourcePath := "anxcloud_ip_address." + resourceName

	prefixID := "0d82d7fdbb804e7fab445c3f85ce7e90"
	ipAddress := "10.244.2.19"
	role := "Reserved"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccAnxCloudIPAddressDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAnxCloudIPAddress(resourceName, prefixID, ipAddress, role),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "network_prefix_id", prefixID),
					resource.TestCheckResourceAttr(resourcePath, "role", role),
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

func TestAccAnxCloudIPAddressTags(t *testing.T) {
	environment.SkipIfNoEnvironment(t)
	envInfo := environment.GetEnvInfo(t)

	prefixID := envInfo.Prefix.ID
	ipAddress := envInfo.Prefix.GetNextIP()
	role := "Default"

	tpl := fmt.Sprintf(`
	resource "anxcloud_ip_address" "foo" {
		network_prefix_id   = "%s"
		address = "%s"
		role = "%s"
		description_customer = "tf-acc-tags"
		
		%%s // tags
	}`, prefixID, ipAddress, role)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccAnxCloudIPAddressDestroy,
		Steps: testAccAnxCloudCommonResourceTagTestSteps(
			tpl, "anxcloud_ip_address.foo",
		),
	})
}

func testAccAnxCloudIPAddress(resourceName, prefixID, ipAddress, role string) string {
	return fmt.Sprintf(`
	resource "anxcloud_ip_address" "%s" {
		network_prefix_id   = "%s"
		address = "%s"
		role = "%s"
	}
	`, resourceName, prefixID, ipAddress, role)
}

func testAccAnxCloudIPAddressDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(providerContext).legacyClient
	a := address.NewAPI(c)
	ctx := context.Background()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "anxcloud_ip_address" {
			continue
		}

		if rs.Primary.ID == "" {
			return nil
		}

		info, err := a.Get(ctx, rs.Primary.ID)
		if err != nil {
			if err := handleNotFoundError(err); err != nil {
				return err
			}
			return nil
		}
		if info.Status != ipAddressStatusDeleted {
			return fmt.Errorf("ip address '%s' exists", info.ID)
		}
	}

	return nil
}
