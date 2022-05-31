package anxcloud

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/anexia-it/terraform-provider-anxcloud/anxcloud/testutils/environment"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAnxCloudIPAddressDataSource(t *testing.T) {
	environment.SkipIfNoEnvironment(t)

	envInfo := environment.GetEnvInfo(t)

	testResources := testAccAnxCloudIPAddressDataSourceResources(envInfo)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			// Get IPv4 address by it's id
			{
				Config: testResources + `
				data "anxcloud_ip_address" "test_id" {
					id         = anxcloud_ip_address.test_v4_x_x_x_5.id
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.anxcloud_ip_address.test_id", "vlan_id", "anxcloud_ip_address.test_v4_x_x_x_5", "vlan_id"),
				),
			},

			// Get IPv4 address by it's "address"
			{
				Config: testResources + `
				data "anxcloud_ip_address" "test_address" {
					address    = anxcloud_ip_address.test_v4_x_x_x_5.address
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.anxcloud_ip_address.test_address", "vlan_id", "anxcloud_ip_address.test_v4_x_x_x_5", "vlan_id"),
				),
			},

			// Get another IPv4 address with same address-name-prefix as previous (to test search-match loop)
			{
				Config: testResources + `
				data "anxcloud_ip_address" "test_address" {
					address    = anxcloud_ip_address.test_v4_x_x_x_50.address
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.anxcloud_ip_address.test_address", "id", "anxcloud_ip_address.test_v4_x_x_x_50", "id"),
				),
			},

			// Get IPv4 address by it's "address" and network_prefix_id
			{
				Config: testResources + `
				data "anxcloud_ip_address" "test_address" {
					address           = anxcloud_ip_address.test_v4_x_x_x_5.address
					network_prefix_id = anxcloud_ip_address.test_v4_x_x_x_5.network_prefix_id
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.anxcloud_ip_address.test_address", "vlan_id", "anxcloud_ip_address.test_v4_x_x_x_5", "vlan_id"),
				),
			},

			// Get IPv4 address by it's "address" and vlan_id
			{
				Config: testResources + `
				data "anxcloud_ip_address" "test_address" {
					address = anxcloud_ip_address.test_v4_x_x_x_5.address
					vlan_id = anxcloud_ip_address.test_v4_x_x_x_5.vlan_id
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.anxcloud_ip_address.test_address", "vlan_id", "anxcloud_ip_address.test_v4_x_x_x_5", "vlan_id"),
				),
			},

			// Get IPv6 address by it's shortened "address"
			{
				Config: testResources + `
				data "anxcloud_ip_address" "test_v6_short" {
					address = anxcloud_ip_address.test_v6.address
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.anxcloud_ip_address.test_v6_short", "network_prefix_id", "anxcloud_ip_address.test_v6", "network_prefix_id"),
				),
			},

			// Test not existing IPv4 address handling
			{
				Config: `
				data "anxcloud_ip_address" "test" {
					address = "1.1.1.1"
				}
				`,
				ExpectError: regexp.MustCompile("IP address was not found."),
			},

			// Test invalid IP address
			{
				Config: `
				data "anxcloud_ip_address" "test" {
					address = "1.1.1.1.1"
				}
				`,
				ExpectError: regexp.MustCompile("Failed to parse IP address."),
			},
		},
	})
}

func testAccAnxCloudIPAddressDataSourceResources(envInfo environment.Info) string {
	return fmt.Sprintf(`
	resource "anxcloud_vlan" "test" {
		location_id     = "%[1]s"
		vm_provisioning = true

		description_customer = "tf-acc-test"
	}

	resource "anxcloud_network_prefix" "test_v4" {
		vlan_id     = anxcloud_vlan.test.id
		location_id = "%[1]s"
		ip_version  = 4
		type        = 1
		netmask     = 24

		description_customer = "tf-acc-test"
	}

	resource "anxcloud_network_prefix" "test_v6" {
		vlan_id     = anxcloud_vlan.test.id
		location_id = "%[1]s"
		ip_version  = 6
		type        = 1
		netmask     = 64

		description_customer = "tf-acc-test"
	}

	resource "anxcloud_ip_address" "test_v4_x_x_x_5" {
		address           = cidrhost(anxcloud_network_prefix.test_v4.cidr, 5)
		network_prefix_id = anxcloud_network_prefix.test_v4.id
	}

	resource "anxcloud_ip_address" "test_v4_x_x_x_50" {
		address           = cidrhost(anxcloud_network_prefix.test_v4.cidr, 50)
		network_prefix_id = anxcloud_network_prefix.test_v4.id
	}

	resource "anxcloud_ip_address" "test_v6" {
		address           = cidrhost(anxcloud_network_prefix.test_v6.cidr, 5)
		network_prefix_id = anxcloud_network_prefix.test_v6.id
	}
	`, envInfo.Location)
}
