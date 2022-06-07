package anxcloud

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAnxCloudCoreLocationDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAnxCloudCoreLocationDataSource("anx04", "ANX04"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.anxcloud_core_location.anx04", "identifier", "52b5f6b2fd3a4a7eaaedf1a7c019e9ea"),
				),
			},
			{
				Config:      testAccAnxCloudCoreLocationDataSource("test", "invalid-code"),
				ExpectError: regexp.MustCompile("No Location found for: invalid-code"),
			},
			{
				Config:      `data "anxcloud_core_location" "test" {}`,
				ExpectError: regexp.MustCompile("location data-source requires code argument"),
			},
		},
	})
}

func testAccAnxCloudCoreLocationDataSource(dataSourceName, code string) string {
	return fmt.Sprintf(`
	data "anxcloud_core_location" "%s" {
		code = "%s"
	}
	`, dataSourceName, code)
}
