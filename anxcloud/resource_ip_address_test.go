package anxcloud

import (
	"context"
	"fmt"
	"github.com/anexia-it/terraform-provider-anxcloud/anxcloud/testutils/components"
	"github.com/anexia-it/terraform-provider-anxcloud/anxcloud/testutils/environment"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"go.anx.io/go-anxcloud/pkg/client"
	"go.anx.io/go-anxcloud/pkg/ipam/address"
)

func TestAccAnxCloudIPAddress(t *testing.T) {
	environment.NewContext(context.Background(), environment.Info{
		Location: "52b5f6b2fd3a4a7eaaedf1a7c019e9ea",
		VlanID:   "1f45f9a5275845fa8249d59af465ce80",
	})
	resourceName := "acc_test"
	resourcePath := "anxcloud_ip_address." + resourceName

	var prefixID string
	ipAddress := "10.244.2.18"
	role := "Default"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			prefix, err := components.CreateTestPrefix(context.Background())
			if err != nil {
				t.Fatal("unable to create prefix for testing", err)
			}
			prefixID = prefix
		},
		ProviderFactories: testAccProviders,
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

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
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
	c := testAccProvider.Meta().(client.Client)
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
