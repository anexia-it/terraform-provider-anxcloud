package anxcloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAnxCloudVLANsDataSource(t *testing.T) {
	resourceName := "acc_test"
	resourcePath := "data.anxcloud_vlans." + resourceName

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAnxCloudVLANsDataSource(resourceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "page", "1"),
					resource.TestCheckResourceAttr(resourcePath, "limit", "1000"),
					resource.TestCheckResourceAttr(resourcePath, "search", ""),
					testAccAnxCloudVLANsDataSourceExists(resourcePath),
				),
			},
		},
	})
}

func testAccAnxCloudVLANsDataSource(resourceName string) string {
	return fmt.Sprintf(`
	data "anxcloud_vlans" "%s" {}
	`, resourceName)
}

func testAccAnxCloudVLANsDataSourceExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("vlans not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("vlans id not set")
		}

		if len(rs.Primary.Attributes) < 1 {
			return fmt.Errorf("not found vlans")
		}

		return nil
	}
}
