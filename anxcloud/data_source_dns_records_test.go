package anxcloud

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAnxCloudDNSRecordsDataSource(t *testing.T) {
	resourceName := "acc_test"
	resourcePath := "data.anxcloud_dns_records." + resourceName

	zoneName := "ake-dev.go-sdk.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAnxCloudDNSRecordsDataSource(resourceName, zoneName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "zone_name", zoneName),
					testAccAnxCloudDNSRecordsDataSourceExists(resourcePath),
				),
			},
		},
	})
}

func testAccAnxCloudDNSRecordsDataSource(resourceName, zoneName string) string {
	return fmt.Sprintf(`
	data "anxcloud_dns_records" "%s" { 
		zone_name = "%s" 
	}
`, resourceName, zoneName)
}

func testAccAnxCloudDNSRecordsDataSourceExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("dns records not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("dns records id not set")
		}

		if len(rs.Primary.Attributes) < 1 {
			return errors.New("dns records not found")
		}
		return nil
	}
}
