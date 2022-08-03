package anxcloud

import (
	"context"
	"fmt"
	"testing"

	"github.com/anexia-it/terraform-provider-anxcloud/anxcloud/testutils/environment"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	lbaasv1 "go.anx.io/go-anxcloud/pkg/apis/lbaas/v1"
)

func TestAccAnxCloudLBaaSLoadBalancer(t *testing.T) {
	environment.SkipIfNoEnvironment(t)
	envInfo := environment.GetEnvInfo(t)

	resourcePath := "anxcloud_lbaas_loadbalancer.foo"
	loadBalancerName := envInfo.TestRunName

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAnxCloudLBaaSLoadBalancer(loadBalancerName, "foo.test"),
				Check:  testAccAnxCloudLBaaSLoadBalancerExists(resourcePath, loadBalancerName, "foo.test"),
			},
			{
				Config: testAccAnxCloudLBaaSLoadBalancer(loadBalancerName, "bar.test"),
				Check:  testAccAnxCloudLBaaSLoadBalancerExists(resourcePath, loadBalancerName, "bar.test"),
			},
			{
				ResourceName:      resourcePath,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAnxCloudLBaaSLoadBalancer(name, ipAddress string) string {
	return fmt.Sprintf(`
	resource "anxcloud_lbaas_loadbalancer" "foo" {
		name = "%s"
		ip_address = "%s"
		tags = ["tf_acc_test"]
	}
	`, name, ipAddress)
}

func testAccAnxCloudLBaaSLoadBalancerExists(n, name, ipAddress string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found in state: %s", n)
		}

		a := testAccProvider.Meta().(providerContext).api

		loadBalancer := lbaasv1.LoadBalancer{Identifier: rs.Primary.ID}
		if err := a.Get(context.Background(), &loadBalancer); err != nil {
			return fmt.Errorf("failed to retrieve LoadBalancer: %w", err)
		}

		if loadBalancer.Name != name {
			return fmt.Errorf("remote LoadBalancer name %q does not match %q", loadBalancer.Name, name)
		}

		if loadBalancer.IpAddress != ipAddress {
			return fmt.Errorf("remote LoadBalancer IP address %q does not match %q", loadBalancer.IpAddress, ipAddress)
		}

		return nil
	}
}
