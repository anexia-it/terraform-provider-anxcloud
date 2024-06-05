package anxcloud

import (
	"context"
	"fmt"
	"net/netip"
	"testing"
	"time"

	"github.com/anexia-it/terraform-provider-anxcloud/anxcloud/testutils/environment"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"go.anx.io/go-anxcloud/pkg/ipam/address"
)

func TestAccAnxCloudIPAddress(t *testing.T) {
	environment.SkipIfNoEnvironment(t)
	resourceName := "acc_test"
	resourcePath := "anxcloud_ip_address." + resourceName

	envInfo := environment.GetEnvInfo(t)

	prefixID := envInfo.Prefix.ID
	ipAddress := envInfo.Prefix.GetNextIP()
	role := "Default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccAnxCloudIPAddressDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAnxCloudIPAddress(resourceName, prefixID, ipAddress, role),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "network_prefix_id", prefixID),
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

func TestAccAnxCloudIPAddressReserved(t *testing.T) {
	resourceName := "acc_test"
	resourcePath := "anxcloud_ip_address." + resourceName

	prefixID := "24e9db909a714dc6ac6ccc19107d410c"
	ipAddress := "10.244.6.2"
	role := "Reserved"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccAnxCloudIPAddressDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAnxCloudIPAddress(resourceName, prefixID, ipAddress, role),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "network_prefix_id", prefixID),
					resource.TestCheckResourceAttr(resourcePath, "role", role),
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

func TestAccAnxCloudIPAddressReserveAvailable(t *testing.T) {
	environment.SkipIfNoEnvironment(t)
	envInfo := environment.GetEnvInfo(t)

	var (
		v4TestPrefix netip.Prefix
		v6TestPrefix netip.Prefix
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					locals {
						test_run_name = %q
					}

					data "anxcloud_core_location" "anx04" {
					  code = "ANX04"
					}

					resource "anxcloud_vlan" "foo" {
					  location_id          = data.anxcloud_core_location.anx04.id
					  vm_provisioning      = true
					  description_customer = "tf-acc-test ${local.test_run_name} anxcloud_ip_address reserve"
					}

					resource "anxcloud_network_prefix" "v4" {
					  location_id          = data.anxcloud_core_location.anx04.id
					  netmask              = 28
					  vlan_id              = anxcloud_vlan.foo.id
					  ip_version           = 4
					  type                 = 1
					  description_customer = "tf-acc-test ${local.test_run_name} anxcloud_ip_address reserve"
					  create_empty         = true
					}

					resource "anxcloud_network_prefix" "v6" {
					  location_id          = data.anxcloud_core_location.anx04.id
					  netmask              = 64
					  vlan_id              = anxcloud_vlan.foo.id
					  ip_version           = 6
					  type                 = 1
					  description_customer = "tf-acc-test ${local.test_run_name} anxcloud_ip_address reserve"
					  create_empty         = true
					}

					resource "anxcloud_ip_address" "v4version" {
					  vlan_id = anxcloud_vlan.foo.id
					  version = 4
					  reservation_period_seconds = 60

					  depends_on = [
						anxcloud_network_prefix.v4,
					  ]
					}

					resource "anxcloud_ip_address" "v6version" {
					  vlan_id = anxcloud_vlan.foo.id
					  version = 6
					  reservation_period_seconds = 60

					  depends_on = [
						anxcloud_network_prefix.v6,
					  ]
					}

					resource "anxcloud_ip_address" "v4prefix" {
					  vlan_id           = anxcloud_vlan.foo.id
					  network_prefix_id = anxcloud_network_prefix.v4.id
					  reservation_period_seconds = 60
					}

					resource "anxcloud_ip_address" "v6prefix" {
					  vlan_id           = anxcloud_vlan.foo.id
					  network_prefix_id = anxcloud_network_prefix.v6.id
					  reservation_period_seconds = 60
					}

					resource "anxcloud_ip_address" "anyprefixorversion" {
					  vlan_id = anxcloud_vlan.foo.id
					  reservation_period_seconds = 60

					  depends_on = [
						anxcloud_network_prefix.v4,
						anxcloud_network_prefix.v6,
					  ]
					}
				`, envInfo.TestRunName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrWith("anxcloud_network_prefix.v4", "cidr", func(value string) error {
						v4TestPrefix = netip.MustParsePrefix(value)
						return nil
					}),
					resource.TestCheckResourceAttrWith("anxcloud_network_prefix.v6", "cidr", func(value string) error {
						v6TestPrefix = netip.MustParsePrefix(value)
						return nil
					}),
					resource.TestCheckResourceAttrWith("anxcloud_ip_address.v4version", "address", func(value string) error {
						if !netip.MustParseAddr(value).Is4() {
							return fmt.Errorf("not a v4 address")
						}
						return nil
					}),
					resource.TestCheckResourceAttrWith("anxcloud_ip_address.v6version", "address", func(value string) error {
						if !netip.MustParseAddr(value).Is6() {
							return fmt.Errorf("not a v6 address")
						}
						return nil
					}),
					resource.TestCheckResourceAttrWith("anxcloud_ip_address.v4prefix", "address", func(value string) error {
						if !v4TestPrefix.Contains(netip.MustParseAddr(value)) {
							return fmt.Errorf("address not in v4 test prefix")
						}
						return nil
					}),
					resource.TestCheckResourceAttrWith("anxcloud_ip_address.v6prefix", "address", func(value string) error {
						if !v6TestPrefix.Contains(netip.MustParseAddr(value)) {
							return fmt.Errorf("address not in v6 test prefix")
						}
						return nil
					}),
				),
			},
			// wait for the reservations to expire before tearing down prefixes
			{PreConfig: func() { time.Sleep(10 * time.Minute) }, Config: "# empty config"},
		},
	})
}

func TestAccAnxCloudIPAddressTags(t *testing.T) {
	environment.SkipIfNoEnvironment(t)
	envInfo := environment.GetEnvInfo(t)

	prefixID := envInfo.Prefix.ID
	ipAddress := envInfo.Prefix.GetNextIP()
	role := "Default"

	tpl := fmt.Sprintf(`
	resource "anxcloud_ip_address" "foo" {
		network_prefix_id   = "%s"
		address = "%s"
		role = "%s"
		description_customer = "tf-acc-tags"
		
		%%s // tags
	}`, prefixID, ipAddress, role)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccAnxCloudIPAddressDestroy,
		Steps: testAccAnxCloudCommonResourceTagTestSteps(
			tpl, "anxcloud_ip_address.foo",
		),
	})
}

func testAccAnxCloudIPAddress(resourceName, prefixID, ipAddress, role string) string {
	return fmt.Sprintf(`
	resource "anxcloud_ip_address" "%s" {
		network_prefix_id   = "%s"
		address = "%s"
		role = "%s"
	}
	`, resourceName, prefixID, ipAddress, role)
}

func testAccAnxCloudIPAddressDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(providerContext).legacyClient
	a := address.NewAPI(c)
	ctx := context.Background()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "anxcloud_ip_address" {
			continue
		}

		if rs.Primary.ID == "" {
			return nil
		}

		info, err := a.Get(ctx, rs.Primary.ID)
		if err != nil {
			if err := handleNotFoundError(err); err != nil {
				return err
			}
			return nil
		}
		if info.Status != ipAddressStatusDeleted {
			return fmt.Errorf("ip address '%s' exists", info.ID)
		}
	}

	return nil
}
