package anxcloud

import (
	"context"
	"fmt"
	"testing"

	"github.com/anexia-it/terraform-provider-anxcloud/anxcloud/testutils/environment"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"go.anx.io/go-anxcloud/pkg/ipam/prefix"
)

func TestAccAnxCloudNetworkPrefix(t *testing.T) {
	environment.SkipIfNoEnvironment(t)

	resourceName := "acc_test"
	resourcePath := "anxcloud_network_prefix." + resourceName

	envInfo := environment.GetEnvInfo(t)
	locationID := envInfo.Location
	customerDescription := "network prefix acceptance tests: " + envInfo.TestRunName
	customerDescriptionUpdate := "network prefix acceptance tests update"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAnxCloudNetworkPrefixDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAnxCloudNetworkPrefix(resourceName, locationID, customerDescription, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "location_id", locationID),
					resource.TestCheckResourceAttr(resourcePath, "description_customer", customerDescription),
					resource.TestCheckResourceAttr(resourcePath, "type", "1"),
					resource.TestCheckResourceAttr(resourcePath, "create_empty", "true"),
					testAccAnxCloudNetworkPrefixExists(resourcePath, customerDescription),
				),
			},
			{
				Config: testAccAnxCloudNetworkPrefix(resourceName, locationID, customerDescriptionUpdate, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "location_id", locationID),
					resource.TestCheckResourceAttr(resourcePath, "description_customer", customerDescriptionUpdate),
					resource.TestCheckResourceAttr(resourcePath, "type", "1"),
					resource.TestCheckResourceAttr(resourcePath, "create_empty", "true"),
					testAccAnxCloudNetworkPrefixExists(resourcePath, customerDescriptionUpdate),
				),
			},
			{
				Config: testAccAnxCloudNetworkPrefix(resourceName, locationID, customerDescriptionUpdate, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "location_id", locationID),
					resource.TestCheckResourceAttr(resourcePath, "description_customer", customerDescriptionUpdate),
					resource.TestCheckResourceAttr(resourcePath, "type", "1"),
					resource.TestCheckResourceAttr(resourcePath, "create_empty", "false"),
					testAccAnxCloudNetworkPrefixExists(resourcePath, customerDescriptionUpdate),
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

func TestAccAnxCloudNetworkPrefixTags(t *testing.T) {
	environment.SkipIfNoEnvironment(t)
	envInfo := environment.GetEnvInfo(t)

	tpl := fmt.Sprintf(`
	resource "anxcloud_network_prefix" "foo" {
		vlan_id     = "%s"
		location_id = "%s"
		ip_version  = 4
		type 				= 1
		netmask 		= 31
		description_customer = "tf-acc-tags"

		%%s // tags
	}`, envInfo.VlanID, envInfo.Location)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAnxCloudNetworkPrefixDestroy,
		Steps: testAccAnxCloudCommonResourceTagTestSteps(
			tpl, "anxcloud_network_prefix.foo",
		),
	})
}

func testAccCheckAnxCloudNetworkPrefixDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(providerContext).legacyClient
	p := prefix.NewAPI(c)
	ctx := context.Background()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "anxcloud_network_prefix" {
			continue
		}

		if rs.Primary.ID == "" {
			return nil
		}

		info, err := p.Get(ctx, rs.Primary.ID)
		if err != nil {
			if err := handleNotFoundError(err); err != nil {
				return err
			}
			return nil
		}
		if info.Status != prefixStatusDeleted {
			return fmt.Errorf("vlan '%s' exists", info.ID)
		}
	}

	return nil
}

func testAccAnxCloudNetworkPrefix(resourceName, locationID, customerDescription string, createEmpty bool) string {
	return fmt.Sprintf(`
	resource "anxcloud_network_prefix" "%s" {
		location_id   = "%s"
		vlan_id = "00a239d617504e4ab49122efe0d27657"
		ip_version = 4
		netmask = 30
		type = 1
		description_customer = "%s"
		create_empty = %v
	}
	`, resourceName, locationID, customerDescription, createEmpty)
}

func testAccAnxCloudNetworkPrefixExists(n string, expectedCustomerDescription string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		c := testAccProvider.Meta().(providerContext).legacyClient
		p := prefix.NewAPI(c)
		ctx := context.Background()

		if !ok {
			return fmt.Errorf("network prefix not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("network prefix id not set")
		}

		i, err := p.Get(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		if i.Status != prefixStatusActive {
			return fmt.Errorf("network prefix found but it is not in the expected state '%s': %s", prefixStatusActive, i.Status)
		}

		if i.CustomerDescription != expectedCustomerDescription {
			return fmt.Errorf("customer description is different than expected '%s': '%s'", i.CustomerDescription, expectedCustomerDescription)
		}

		return nil
	}
}
