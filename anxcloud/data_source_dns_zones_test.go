package anxcloud

import (
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"testing"
)

func TestAccAnxCloudDnsZonessDataSource(t *testing.T) {
	resourceName := "acc_test"
	resourcePath := "data.anxcloud_zones_records." + resourceName

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAnxCloudDnsZonesDataSource(resourceName),
				Check: resource.ComposeTestCheckFunc(
					//resource.TestCheckResourceAttr(resourcePath, "zone_name", zoneName),
					testAccAnxCloudDnsZonesDataSourceExists(resourcePath),
				),
			},
		},
	})
}

func testAccAnxCloudDnsZonesDataSource(resourceName string) string {
	return fmt.Sprintf(`
	data "anxcloud_dns_zones" "%s" {}
`, resourceName)
}

func testAccAnxCloudDnsZonesDataSourceExists(n string) resource.TestCheckFunc {
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
