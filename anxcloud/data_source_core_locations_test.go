package anxcloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAnxCloudCoreLocationsDataSource(t *testing.T) {
	resourceName := "acc_test"
	resourcePath := "data.anxcloud_core_locations." + resourceName

	search := "IE"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAnxCloudCoreLocationsDataSource(resourceName, search),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "search", search),
					testAccAnxCloudCoreLocationsDataSourceExists(resourcePath),
				),
			},
		},
	})
}

func testAccAnxCloudCoreLocationsDataSource(resourceName, search string) string {
	return fmt.Sprintf(`
	data "anxcloud_core_locations" "%s" {
		search = "%s"
	}
	`, resourceName, search)
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
