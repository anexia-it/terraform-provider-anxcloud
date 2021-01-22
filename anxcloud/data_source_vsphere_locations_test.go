package anxcloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAnxCloudLocationDataSource(t *testing.T) {
	resourceName := "acc_locations_test"
	resourcePath := "data.anxcloud_locations." + resourceName

	page := "1"
	limit := "50"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAnxCloudLocationDataSource(resourceName, page, limit),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "page", page),
					resource.TestCheckResourceAttr(resourcePath, "limit", limit),
					testAccAnxCloudLocationDataSourceExists(resourcePath),
				),
			},
		},
	})
}

func testAccAnxCloudLocationDataSource(resourceName, page, limit string) string {
	return fmt.Sprintf(`
	data "anxcloud_location" "%s" {
		page = %v
		limit = %v
	}
	`, resourceName, page, limit)
}

func testAccAnxCloudLocationDataSourceExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("locations not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("locations id not set")
		}

		if len(rs.Primary.Attributes) < 1 {
			return fmt.Errorf("not found locations")
		}

		return nil
	}
}
