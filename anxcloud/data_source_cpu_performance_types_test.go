package anxcloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAnxCloudCPUPerformanceTypesDataSource(t *testing.T) {
	resourceName := "acc_cpu_performance_types_test"
	resourcePath := "data.anxcloud_cpu_performance_types." + resourceName

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAnxCloudCPUPerformanceTypesDataSource(resourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccAnxCloudCPUPerformanceTypesDataSourceExists(resourcePath),
				),
			},
		},
	})
}

func testAccAnxCloudCPUPerformanceTypesDataSource(resourceName string) string {
	return fmt.Sprintf(`
	data "anxcloud_tags" "%s" {
	}
	`, resourceName)
}

func testAccAnxCloudCPUPerformanceTypesDataSourceExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("tags not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("tags id not set")
		}

		if len(rs.Primary.Attributes) < 1 {
			return fmt.Errorf("not found tags")
		}

		return nil
	}
}
