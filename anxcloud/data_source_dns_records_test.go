package anxcloud

import (
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"testing"
)

func TestAccAnxCloudDnsRecordsDataSource(t *testing.T) {
	resourceName := "acc_test"
	resourcePath := "data.anxcloud_dns_records." + resourceName

	zoneName := "go-sdk.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAnxCloudDnsRecordsDataSource(resourceName, zoneName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "zone_name", zoneName),
					testAccAnxCloudDnsRecordsDataSourceExists(resourcePath),
				),
			},
		},
	})
}

func testAccAnxCloudDnsRecordsDataSource(resourceName, zoneName string) string {
	return fmt.Sprintf(`
	data "anxcloud_dns_records" "%s" { 
		zone_name = "%s" 
	}
`, resourceName, zoneName)
}

func testAccAnxCloudDnsRecordsDataSourceExists(n string) resource.TestCheckFunc {
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
