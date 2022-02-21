package anxcloud

import (
	"fmt"
	"github.com/anexia-it/terraform-provider-anxcloud/anxcloud/testutils/environment"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAnxCloudDiskTypeDataSource(t *testing.T) {
	environment.SkipIfNoEnvironment(t)

	resourceName := "acc_test"
	resourcePath := "data.anxcloud_disk_types." + resourceName

	locationID := environment.GetEnvInfo(t).Location

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAnxCloudDiskTypeDataSource(resourceName, locationID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "location_id", locationID),
					resource.TestCheckResourceAttr(resourcePath, "id", locationID),
					testAccAnxCloudDiskTypeDataSourceExists(resourcePath),
				),
			},
		},
	})
}

func testAccAnxCloudDiskTypeDataSource(resourceName, locationID string) string {
	return fmt.Sprintf(`
	data "anxcloud_disk_types" "%s" {
		location_id   = "%s"
	}
	`, resourceName, locationID)
}

func testAccAnxCloudDiskTypeDataSourceExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("disk types not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("disk types id not set")
		}

		if len(rs.Primary.Attributes) < 1 {
			return fmt.Errorf("not found disk types")
		}

		return nil
	}
}
