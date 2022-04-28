package anxcloud

import (
	"fmt"
	"testing"

	"github.com/anexia-it/terraform-provider-anxcloud/anxcloud/testutils/environment"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"go.anx.io/go-anxcloud/pkg/utils/test"
)

func TestAccAnxCloudDNSRecord(t *testing.T) {
	environment.SkipIfNoEnvironment(t)
	resourceName := "acc_test"

	zoneName := test.RandomHostname() + ".terraform.test"
	recordName := "test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAnxDNSZoneAndRecord(zoneName, resourceName, recordName),
			},
		},
	})
}

func testAccAnxDNSZoneAndRecord(zoneName, resourceName, recordName string) string {
	return fmt.Sprintf(`
	resource "anxcloud_dns_zone" "test_dns_zone" {
		name = "%s"
		is_master = true
		dns_sec_mode = "unvalidated"
		admin_email = "admin@%s"
		refresh = 100
		retry = 100
		expire = 1000
		ttl = 100
	}

	resource "anxcloud_dns_record" "%s" {
		name = "%s"
		zone_name = anxcloud_dns_zone.test_dns_zone.name
		type = "A"
		rdata = "1.1.1.1"
		ttl = 300
	}
	`, zoneName, zoneName, resourceName, recordName)
}
