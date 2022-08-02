package anxcloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAnxCloudIPAddressesDataSource(t *testing.T) {
	resourceName := "acc_test"
	resourcePath := "data.anxcloud_ip_addresses." + resourceName

	search := "10.244"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAnxCloudIPAddressesDataSource(resourceName, search),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "search", search),
					testAccAnxCloudIPAddressesDataSourceExists(resourcePath),
				),
			},
		},
	})
}

func testAccAnxCloudIPAddressesDataSource(resourceName, search string) string {
	return fmt.Sprintf(`
	data "anxcloud_ip_addresses" "%s" {
		search = "%s"
	}
	`, resourceName, search)
}

func testAccAnxCloudIPAddressesDataSourceExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("ip addresses not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("ip addresses id not set")
		}

		if len(rs.Primary.Attributes) < 1 {
			return fmt.Errorf("not found ip addresses")
		}

		return nil
	}
}
