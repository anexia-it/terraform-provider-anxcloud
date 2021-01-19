package anxcloud

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAnxCloudCoreLocationsDataSource(t *testing.T) {
	resourceName := "acc_test"
	resourcePath := "data.anxcloud_core_locations." + resourceName

	page := 1
	limit := 1
	search := "IE"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAnxCloudCoreLocationsDataSource(resourceName, page, limit, search),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "page", strconv.Itoa(page)),
					resource.TestCheckResourceAttr(resourcePath, "limit", strconv.Itoa(limit)),
					resource.TestCheckResourceAttr(resourcePath, "search", search),
					testAccAnxCloudCoreLocationsDataSourceExists(resourcePath),
				),
			},
		},
	})
}

func testAccAnxCloudCoreLocationsDataSource(resourceName string, page, limit int, search string) string {
	return fmt.Sprintf(`
	data "anxcloud_core_locations" "%s" {
		page   = %d
		limit  = %d
		search = "%s"
	}
	`, resourceName, page, limit, search)
}

func testAccAnxCloudCoreLocationsDataSourceExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("core locations not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("core locations id not set")
		}

		if len(rs.Primary.Attributes) < 1 {
			return fmt.Errorf("not found core locations")
		}

		return nil
	}
}
