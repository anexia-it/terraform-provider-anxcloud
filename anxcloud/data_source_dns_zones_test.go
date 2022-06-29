package anxcloud

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAnxCloudDNSZonesDataSource(t *testing.T) {
	resourceName := "acc_test"
	resourcePath := "data.anxcloud_dns_zones." + resourceName

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAnxCloudDNSZonesDataSource(resourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccAnxCloudDNSZonesDataSourceExists(resourcePath),
				),
			},
		},
	})
}

func testAccAnxCloudDNSZonesDataSource(resourceName string) string {
	return fmt.Sprintf(`
	data "anxcloud_dns_zones" "%s" {}
`, resourceName)
}

func testAccAnxCloudDNSZonesDataSourceExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("dns zones not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("dns zones id not set")
		}

		if len(rs.Primary.Attributes) < 1 {
			return errors.New("dns zones not found")
		}
		return nil
	}
}
