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
					resource.TestCheckResourceAttr("anxcloud_dns_record.txt_record", "name", "txt-record"),
					resource.TestCheckResourceAttrSet("anxcloud_dns_record.a_record", "identifier"),
					resource.TestCheckResourceAttrSet("anxcloud_dns_record.txt_record", "identifier"),
				),
			},
			{
				Config: testAccAnxDNSZoneAndRecord(zoneName, 1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("anxcloud_dns_record.a_record", "zone_name", "1-"+zoneName),
					resource.TestCheckResourceAttr("anxcloud_dns_record.a_record", "name", "a-record"),
					resource.TestCheckResourceAttr("anxcloud_dns_record.txt_record", "name", "txt-record"),
					resource.TestCheckResourceAttrSet("anxcloud_dns_record.a_record", "identifier"),
					resource.TestCheckResourceAttrSet("anxcloud_dns_record.txt_record", "identifier"),
				),
			},
		},
	})
}

func TestAccAnxCloudDNSRecordImport(t *testing.T) {
	environment.SkipIfNoEnvironment(t)
	zoneName := test.RandomHostname() + ".terraform.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAnxDNSZoneAndRecordImport(zoneName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("anxcloud_dns_record.import_test", "zone_name", zoneName),
					resource.TestCheckResourceAttr("anxcloud_dns_record.import_test", "name", "import-test"),
					resource.TestCheckResourceAttr("anxcloud_dns_record.import_test", "type", "A"),
					resource.TestCheckResourceAttr("anxcloud_dns_record.import_test", "rdata", "192.168.1.1"),
					resource.TestCheckResourceAttrSet("anxcloud_dns_record.import_test", "identifier"),
				),
			},
			{
				ResourceName:      "anxcloud_dns_record.import_test",
				ImportState:       true,
				ImportStateVerify: true,
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

func testAccAnxDNSZoneAndRecordImport(zoneName string) string {
	return fmt.Sprintf(`
	resource "anxcloud_dns_zone" "import_zone" {
		name = "%s"
		is_master = true
		dns_sec_mode = "unvalidated"
		admin_email = "admin@terraform.test"
		refresh = 100
		retry = 100
		expire = 1000
		ttl = 100
	}

	resource "anxcloud_dns_record" "import_test" {
		name = "import-test"
		zone_name = anxcloud_dns_zone.import_zone.name
		type = "A"
		rdata = "192.168.1.1"
		ttl = 300
	}
	`, zoneName)
}
