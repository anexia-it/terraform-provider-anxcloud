package anxcloud

// import (
// 	"fmt"
// 	"testing"

// 	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
// 	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
// )

// func TestAccAnxCloudTagsDataSource(t *testing.T) {
// 	resourceName := "acc_tags_test"
// 	resourcePath := "data.anxcloud_tags." + resourceName

// 	resource.Test(t, resource.TestCase{
// 		PreCheck:          func() { testAccPreCheck(t) },
// 		ProviderFactories: testAccProviders,
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccAnxCloudTagsDataSource(resourceName),
// 				Check: resource.ComposeTestCheckFunc(
// 					resource.TestCheckResourceAttr(resourcePath, "location_id", locationID),
// 					resource.TestCheckResourceAttr(resourcePath, "id", locationID),
// 					// page and limit
// 					testAccAnxCloudTagsDataSourceExists(resourcePath),
// 				),
// 			},
// 		},
// 	})
// }

// func testAccAnxCloudTagsDataSource(resourceName string) string {
// 	return fmt.Sprintf(`
// 	data "anxcloud_tags" "%s" {
// 	}
// 	`, resourceName)

// 	return fmt.Sprintf(`
// 	data "anxcloud_disk_type" "%s" {
// 		location_id   = "%s"
// 	}
// 	`, resourceName, locationID)
// }

// func testAccAnxCloudTagsDataSourceExists(n string) resource.TestCheckFunc {
// 	return func(s *terraform.State) error {
// 		rs, ok := s.RootModule().Resources[n]

// 		if !ok {
// 			return fmt.Errorf("tags not found: %s", n)
// 		}

// 		if rs.Primary.ID == "" {
// 			return fmt.Errorf("tags id not set")
// 		}

// 		if len(rs.Primary.Attributes) < 1 {
// 			return fmt.Errorf("not found tags")
// 		}

// 		return nil
// 	}
// }
