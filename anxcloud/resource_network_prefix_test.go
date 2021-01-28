package anxcloud

import (
	"context"
	"fmt"
	"testing"

	"github.com/anexia-it/go-anxcloud/pkg/client"
	"github.com/anexia-it/go-anxcloud/pkg/ipam/prefix"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAnxCloudNetworkPrefix(t *testing.T) {
	resourceName := "acc_test"
	resourcePath := "anxcloud_network_prefix." + resourceName

	locationID := "52b5f6b2fd3a4a7eaaedf1a7c019e9ea"
	customerDescription := "network prefix acceptance tests"
	customerDescriptionUpdate := "network prefix acceptance tests update"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckAnxCloudNetworkPrefixDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAnxCloudNetworkPrefix(resourceName, locationID, customerDescription),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "location_id", locationID),
					resource.TestCheckResourceAttr(resourcePath, "vm_provisioning", "true"),
					resource.TestCheckResourceAttr(resourcePath, "description_customer", customerDescription),
					testAccAnxCloudNetworkPrefixExists(resourcePath, customerDescription),
				),
			},
			{
				Config: testAccAnxCloudNetworkPrefix(resourceName, locationID, customerDescriptionUpdate),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "location_id", locationID),
					resource.TestCheckResourceAttr(resourcePath, "vm_provisioning", "true"),
					resource.TestCheckResourceAttr(resourcePath, "description_customer", customerDescriptionUpdate),
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

func testAccCheckAnxCloudNetworkPrefixDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(client.Client)
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

func testAccAnxCloudNetworkPrefix(resourceName, locationID, customerDescription string) string {
	return fmt.Sprintf(`
	resource "anxcloud_network_prefix" "%s" {
		location_id   = "%s"
		vlan_id = "02f39d20ca0f4adfb5032f88dbc26c39"
		ip_version = 4
		netmask = 30
		vm_provisioning = true
		description_customer = "%s"
	}
	`, resourceName, locationID, customerDescription)
}

func testAccAnxCloudNetworkPrefixExists(n string, expectedCustomerDescription string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		c := testAccProvider.Meta().(client.Client)
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
