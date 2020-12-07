package anxcloud

import (
	"context"
	"fmt"
	"testing"

	"github.com/anexia-it/go-anxcloud/pkg/client"
	"github.com/anexia-it/go-anxcloud/pkg/ipam/address"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAnxCloudIPAddress(t *testing.T) {
	resourceName := "acc_test"
	resourcePath := "anxcloud_ip_address." + resourceName

	prefixID := "7545899235004092b2af15a64419b8c5"
	ipAddress := "185.228.148.118"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccAnxCloudIPAddressDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAnxCloudIPAddress(resourceName, prefixID, ipAddress),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "network_prefix_id", prefixID),
				),
			},
		},
	})
}

func testAccAnxCloudIPAddress(resourceName, prefixID, ipAddress string) string {
	return fmt.Sprintf(`
	resource "anxcloud_ip_address" "%s" {
		network_prefix_id   = "%s"
		address = "%s"
	}
	`, resourceName, prefixID, ipAddress)
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
