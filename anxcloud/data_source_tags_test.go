package anxcloud

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAnxCloudTagsDataSource(t *testing.T) {
	resourceName := "acc_tags_test"
	resourcePath := "data.anxcloud_tags." + resourceName

	page := 1
	limit := 10
	query := "test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAnxCloudTagsDataSource(resourceName, page, limit, query),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "page", strconv.Itoa(page)),
					resource.TestCheckResourceAttr(resourcePath, "limit", strconv.Itoa(limit)),
					resource.TestCheckResourceAttr(resourcePath, "query", query),
					testAccAnxCloudTagsDataSourceExists(resourcePath),
				),
			},
		},
	})
}

func testAccAnxCloudTagsDataSource(resourceName string, page, limit int, query string) string {
	return fmt.Sprintf(`
	data "anxcloud_tags" "%s" {
		page = %d
		limit = %d
		query = "%s"
		sort_ascending = true
	}
	`, resourceName, page, limit, query)
}

func testAccAnxCloudTagsDataSourceExists(n string) resource.TestCheckFunc {
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
