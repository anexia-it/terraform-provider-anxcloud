package anxcloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAnxCloudNICTypesDataSource(t *testing.T) {
	resourceName := "acc_test"
	resourcePath := "data.anxcloud_nic_types." + resourceName

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAnxCloudNICTypesDataSource(resourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccAnxCloudNICTypesDataSourceExists(resourcePath),
				),
			},
		},
	})
}

func testAccAnxCloudNICTypesDataSource(resourceName string) string {
	return fmt.Sprintf(`
	data "anxcloud_nic_types" "%s" {}
	`, resourceName)
}

func testAccAnxCloudNICTypesDataSourceExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("nic types not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("nic types id not set")
		}

		if len(rs.Primary.Attributes) < 1 {
			return fmt.Errorf("not found nic types")
		}

		return nil
	}
}
