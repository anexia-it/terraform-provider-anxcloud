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
	zoneName := test.RandomHostname() + ".terraform.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAnxDNSZoneAndRecord(zoneName, 0),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("anxcloud_dns_record.a_record", "zone_name", "0-"+zoneName),
					resource.TestCheckResourceAttr("anxcloud_dns_record.a_record", "name", "a-record"),
					resource.TestCheckResourceAttr("anxcloud_dns_record.a_record", "comment", "testcomment"),
					resource.TestCheckResourceAttr("anxcloud_dns_record.txt_record", "name", "txt-record"),
				),
			},
			{
				// in-place update: rdata, ttl and comment change without recreating the records
				Config: testAccAnxDNSZoneAndRecordUpdated(zoneName, 0),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("anxcloud_dns_record.a_record", "zone_name", "0-"+zoneName),
					resource.TestCheckResourceAttr("anxcloud_dns_record.a_record", "name", "a-record"),
					resource.TestCheckResourceAttr("anxcloud_dns_record.a_record", "rdata", "2.2.2.2"),
					resource.TestCheckResourceAttr("anxcloud_dns_record.a_record", "ttl", "600"),
					resource.TestCheckResourceAttr("anxcloud_dns_record.a_record", "comment", "updated testcomment"),
					resource.TestCheckResourceAttr("anxcloud_dns_record.txt_record", "name", "txt-record"),
					resource.TestCheckResourceAttr("anxcloud_dns_record.txt_record", "rdata", "hello update"),
				),
			},
			{
				Config: testAccAnxDNSZoneAndRecord(zoneName, 1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("anxcloud_dns_record.a_record", "zone_name", "1-"+zoneName),
					resource.TestCheckResourceAttr("anxcloud_dns_record.a_record", "name", "a-record"),
					resource.TestCheckResourceAttr("anxcloud_dns_record.txt_record", "name", "txt-record"),
				),
			},
		},
	})
}

func testAccAnxDNSZoneAndRecord(zoneNameSuffix string, recordsZoneIndex uint) string {
	return fmt.Sprintf(`
	resource "anxcloud_dns_zone" "test_dns_zones" {
		count = 2
		name = "${count.index}-%s"
		is_master = true
		dns_sec_mode = "unvalidated"
		admin_email = "admin@terraform.test"
		refresh = 100
		retry = 100
		expire = 1000
		ttl = 100
	}

	resource "anxcloud_dns_record" "a_record" {
		name = "a-record"
		zone_name = anxcloud_dns_zone.test_dns_zones[%[2]d].name
		type = "A"
		rdata = "1.1.1.1"
		comment = "testcomment"
		ttl = 300
	}

	resource "anxcloud_dns_record" "txt_record" {
		name = "txt-record"
		zone_name = anxcloud_dns_zone.test_dns_zones[%[2]d].name
		type = "TXT"
		rdata = "hello world"
		ttl = 300
	}
	`, zoneNameSuffix, recordsZoneIndex)
}

func testAccAnxDNSZoneAndRecordUpdated(zoneNameSuffix string, recordsZoneIndex uint) string {
	return fmt.Sprintf(`
	resource "anxcloud_dns_zone" "test_dns_zones" {
		count = 2
		name = "${count.index}-%s"
		is_master = true
		dns_sec_mode = "unvalidated"
		admin_email = "admin@terraform.test"
		refresh = 100
		retry = 100
		expire = 1000
		ttl = 100
	}

	resource "anxcloud_dns_record" "a_record" {
		name = "a-record"
		zone_name = anxcloud_dns_zone.test_dns_zones[%[2]d].name
		type = "A"
		rdata = "2.2.2.2"
		comment = "updated testcomment"
		ttl = 600
	}

	resource "anxcloud_dns_record" "txt_record" {
		name = "txt-record"
		zone_name = anxcloud_dns_zone.test_dns_zones[%[2]d].name
		type = "TXT"
		rdata = "hello update"
		ttl = 300
	}
	`, zoneNameSuffix, recordsZoneIndex)
}
