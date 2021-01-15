package anxcloud

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAnxCloudIPAddressesDataSource(t *testing.T) {
	resourceName := "acc_test"
	resourcePath := "data.anxcloud_ip_addresses." + resourceName

	page := 1
	limit := 1
	search := "10.244"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAnxCloudIPAddressesDataSource(resourceName, page, limit, search),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "page", strconv.Itoa(page)),
					resource.TestCheckResourceAttr(resourcePath, "limit", strconv.Itoa(limit)),
					resource.TestCheckResourceAttr(resourcePath, "search", search),
					testAccAnxCloudIPAddressesDataSourceExists(resourcePath),
				),
			},
		},
	})
}

func testAccAnxCloudIPAddressesDataSource(resourceName string, page, limit int, search string) string {
	return fmt.Sprintf(`
	data "anxcloud_ip_addresses" "%s" {
		page   = %d
		limit  = %d
		search = "%s"
	}
	`, resourceName, page, limit, search)
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
