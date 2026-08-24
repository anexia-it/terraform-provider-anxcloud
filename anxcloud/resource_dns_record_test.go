package anxcloud

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/anexia-it/terraform-provider-anxcloud/anxcloud/testutils/environment"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.anx.io/go-anxcloud/pkg/utils/test"
)

func TestDNSRecordUpdateRequestContract(t *testing.T) {
	provider := kubernetesTestProviderContext(t, func(r *http.Request) (*http.Response, error) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "/api/clouddns/v1/zone.json/example.test/records/record-id", r.URL.Path)

		var body map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, "a-record", body["name"])
		assert.Equal(t, "A", body["type"])
		assert.Equal(t, "2.2.2.2", body["rdata"])
		assert.Equal(t, "default", body["region"])
		assert.Equal(t, float64(600), body["ttl"])
		assert.Equal(t, "updated testcomment", body["comment"])
		assert.NotContains(t, body, "identifier")
		assert.NotContains(t, body, "zone_name")
		assert.NotContains(t, body, "immutable")

		return kubernetesTestResponse(t, http.StatusOK, map[string]any{}), nil
	})

	comment := "updated testcomment"
	update := dnsRecordUpdate{
		zoneName:         "example.test",
		recordIdentifier: "record-id",
		Name:             "a-record",
		Type:             "A",
		RData:            "2.2.2.2",
		Region:           "default",
		TTL:              600,
		Comment:          &comment,
	}

	require.NoError(t, provider.api.Update(context.Background(), &update))
}

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
