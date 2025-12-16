package anxcloud

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/anexia-it/terraform-provider-anxcloud/anxcloud/testutils/environment"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"go.anx.io/go-anxcloud/pkg/utils/test"
)

func TestAccAnxCloudDNSZone(t *testing.T) {
	environment.SkipIfNoEnvironment(t)
	resourceName := "acc_test"
	resourcePath := "anxcloud_dns_zone." + resourceName

	zoneName := test.RandomHostname() + ".terraform.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAnxDNSZone(resourceName, zoneName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "name", zoneName),
				),
			},
			{
				ResourceName:            resourcePath,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"deployment_level", "validation_level"},
			},
			{
				Config:      testAccAnxDNSZone(resourceName, zoneName+"-renamed"),
				ExpectError: regexp.MustCompile("operation not supported"),
			},
		},
	})
}

func TestAccAnxCloudDNSZone_DuplicateDetection(t *testing.T) {
	environment.SkipIfNoEnvironment(t)
	resourceName := "acc_test_duplicate"
	resourcePath := "anxcloud_dns_zone." + resourceName

	zoneName := test.RandomHostname() + ".terraform.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create the zone
			{
				Config: testAccAnxDNSZone(resourceName, zoneName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "name", zoneName),
				),
			},
			// Step 2: Try to create a duplicate zone - should fail at plan time
			{
				Config:      testAccAnxDNSZoneDuplicate(resourceName, zoneName),
				ExpectError: regexp.MustCompile("already exists"),
			},
		},
	})
}

func testAccAnxDNSZone(resourceName, zoneName string) string {
	return fmt.Sprintf(`
	resource "anxcloud_dns_zone" "%s" {
		name = "%[2]s"
		is_master = true
		dns_sec_mode = "unvalidated"
		admin_email = "admin@%[2]s"
		refresh = 100
		retry = 100
		expire = 1000
		ttl = 100
	}
	`, resourceName, zoneName)
}

func testAccAnxDNSZoneDuplicate(resourceName, zoneName string) string {
	return fmt.Sprintf(`
	resource "anxcloud_dns_zone" "%s" {
		name = "%[2]s"
		is_master = true
		dns_sec_mode = "unvalidated"
		admin_email = "admin@%[2]s"
		refresh = 100
		retry = 100
		expire = 1000
		ttl = 100
	}
	
	resource "anxcloud_dns_zone" "%[1]s_duplicate" {
		name = "%[2]s"
		is_master = true
		dns_sec_mode = "unvalidated"
		admin_email = "duplicate@%[2]s"
		refresh = 100
		retry = 100
		expire = 1000
		ttl = 100
	}
	`, resourceName, zoneName)
}
