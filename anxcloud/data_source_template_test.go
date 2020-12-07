package anxcloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAnxCloudTemplateDataSource(t *testing.T) {
	resourceName := "acc_test"
	resourcePath := "data.anxcloud_template." + resourceName

	locationID := "52b5f6b2fd3a4a7eaaedf1a7c019e9ea"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAnxCloudTemplateDataSource(resourceName, locationID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "location_id", locationID),
					resource.TestCheckResourceAttr(resourcePath, "id", locationID),
					testAccAnxCloudTemplateDataSourceExists(resourcePath),
				),
			},
		},
	})
}

func testAccAnxCloudTemplateDataSource(resourceName, locationID string) string {
	return fmt.Sprintf(`
	data "anxcloud_template" "%s" {
		location_id   = "%s"
	}
	`, resourceName, locationID)
}

func testAccAnxCloudTemplateDataSourceExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("template not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("template id not set")
		}

		if len(rs.Primary.Attributes) < 1 {
			return fmt.Errorf("not found templates")
		}

		return nil
	}
}
