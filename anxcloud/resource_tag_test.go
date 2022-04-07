package anxcloud

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"go.anx.io/go-anxcloud/pkg/core/tags"
)

func TestAccAnxCloudTag(t *testing.T) {
	resourceName := "acc_test"
	resourcePath := "anxcloud_tag." + resourceName

	serviceID := "ff543fc08b3149ee9a8c50ee018b15a6"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckAnxCloudTagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAnxCloudTag(resourceName, serviceID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "name", resourceName),
					resource.TestCheckResourceAttr(resourcePath, "service_id", serviceID),
					testAccAnxCloudTagExists(resourcePath),
				),
			},
			{
				ResourceName:      resourcePath,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAnxCloudTagDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(providerContext).legacyClient
	t := tags.NewAPI(c)
	ctx := context.Background()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "anxcloud_network_prefix" {
			continue
		}

		if rs.Primary.ID == "" {
			return nil
		}

		_, err := t.Get(ctx, rs.Primary.ID)
		if err != nil {
			if err := handleNotFoundError(err); err != nil {
				return err
			}
			return nil
		}
	}

	return nil
}

func testAccAnxCloudTag(resourceName, serviceID string) string {
	return fmt.Sprintf(`
	resource "anxcloud_tag" "%s" {
		name = "%s"
		service_id = "%s"
	}
	`, resourceName, resourceName, serviceID)
}

func testAccAnxCloudTagExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		c := testAccProvider.Meta().(providerContext).legacyClient
		t := tags.NewAPI(c)
		ctx := context.Background()

		if !ok {
			return fmt.Errorf("tag not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("tag id not set")
		}

		_, err := t.Get(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		return nil
	}
}
