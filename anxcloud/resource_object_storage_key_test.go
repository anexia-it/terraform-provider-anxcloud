package anxcloud

import (
	"context"
	"fmt"
	"testing"

	"github.com/anexia-it/terraform-provider-anxcloud/anxcloud/testutils/environment"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	objectstoragev2 "go.anx.io/go-anxcloud/pkg/apis/objectstorage/v2"
)

// TestAccAnxCloudObjectStorageKey tests basic object storage key creation
// This is a regression test to ensure the resource correctly handles API Key struct
// without Secret field and properly sets secret_url instead
func TestAccAnxCloudObjectStorageKey(t *testing.T) {
	environment.SkipIfNoEnvironment(t)
	resourceName := "acc_test_key"
	resourcePath := "anxcloud_object_storage_key." + resourceName

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAnxCloudObjectStorageKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAnxCloudObjectStorageKey(resourceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "name", resourceName),
					// Regression: verify secret_url is set (not secret field)
					resource.TestCheckResourceAttrSet(resourcePath, "secret_url"),
					// Verify secret field remains empty as API doesn't return it
					resource.TestCheckResourceAttr(resourcePath, "secret", ""),
					resource.TestCheckResourceAttrSet(resourcePath, "remote_id"),
					testAccAnxCloudObjectStorageKeyExists(resourcePath),
				),
			},
			{
				ResourceName:            resourcePath,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret", "secret_url"}, // secrets are not returned on import
			},
		},
	})
}

func testAccCheckAnxCloudObjectStorageKeyDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(providerContext).api
	ctx := context.Background()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "anxcloud_object_storage_key" {
			continue
		}

		if rs.Primary.ID == "" {
			return nil
		}

		key := objectstoragev2.Key{Identifier: rs.Primary.ID}
		err := c.Get(ctx, &key)
		if err != nil {
			if err := handleNotFoundError(err); err != nil {
				return err
			}
			return nil
		}
	}

	return nil
}

func testAccAnxCloudObjectStorageKey(resourceName string) string {
	return fmt.Sprintf(`
	resource "anxcloud_object_storage_key" "%s" {
		name = "%s"
	}
	`, resourceName, resourceName)
}

func testAccAnxCloudObjectStorageKeyExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		c := testAccProvider.Meta().(providerContext).api
		ctx := context.Background()

		if !ok {
			return fmt.Errorf("object storage key not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("object storage key id not set")
		}

		key := objectstoragev2.Key{Identifier: rs.Primary.ID}
		err := c.Get(ctx, &key)
		if err != nil {
			return err
		}

		// Regression check: verify secret_url is populated by API
		if key.SecretURL == "" {
			return fmt.Errorf("object storage key SecretURL is empty after creation")
		}

		return nil
	}
}
