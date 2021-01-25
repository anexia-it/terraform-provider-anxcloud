package anxcloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAnxCloudVSphereLocationsDataSource(t *testing.T) {
	resourceName := "acc_vsphere_locations_test"
	resourcePath := "data.anxcloud_vsphere_locations." + resourceName

	page := "1"
	limit := "50"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAnxCloudVSphereLocationsDataSource(resourceName, page, limit),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "page", page),
					resource.TestCheckResourceAttr(resourcePath, "limit", limit),
					testAccAnxCloudVSphereLocationsDataSourceExists(resourcePath),
				),
			},
		},
	})
}

func testAccAnxCloudVSphereLocationsDataSource(resourceName, page, limit string) string {
	return fmt.Sprintf(`
	data "anxcloud_vsphere_locations" "%s" {
		page = %v
		limit = %v
	}
	`, resourceName, page, limit)
}

func testAccAnxCloudVSphereLocationsDataSourceExists(n string) resource.TestCheckFunc {
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
