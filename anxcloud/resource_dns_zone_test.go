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

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
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
				Config:      testAccAnxDNSZone(resourceName, "prefix-"+zoneName),
				ExpectError: regexp.MustCompile("operation not supported"),
			},
		},
	})
}

func testAccAnxDNSZone(resourceName, zoneName string) string {
	return fmt.Sprintf(`
	resource "anxcloud_dns_zone" "%s" {
		name = "%s"
		is_master = true
		dns_sec_mode = "unvalidated"
		admin_email = "admin@%s"
		refresh = 100
		retry = 100
		expire = 1000
		ttl = 100
	}
	`, resourceName, zoneName, zoneName)
}
