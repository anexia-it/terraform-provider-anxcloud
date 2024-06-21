package anxcloud

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/anexia-it/terraform-provider-anxcloud/anxcloud/testutils/environment"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAnxCloudVLANDataSource(t *testing.T) {
	environment.SkipIfNoEnvironment(t)

	// The identifier and the name of the VLAN are provided via the environment.
	// However, the providerContext is not given until we are inside the TF tests.
	//
	// And even more unfortunate, the test steps have a static configuration.
	// As a result, vlanID and vlanName have to be known *ahead* of calling
	// resource.ParallelTest. Therefore, we declare them (partially) static here.
	var (
		vlanID   = environment.GetEnvInfo(t).VlanID
		vlanName = "VLAN3286"
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			// expected to fail
			{
				Config:      `data "anxcloud_vlan" "foo" { id = "not_found" }`,
				ExpectError: regexp.MustCompile(`No VLAN with the given identifier "not_found" could be found.`),
			},
			{
				Config:      `data "anxcloud_vlan" "foo" { name = "some VLAN that does not exist" }`,
				ExpectError: regexp.MustCompile(`No VLAN found with the name`),
			},
			{
				Config:      `data "anxcloud_vlan" "foo" {}`,
				ExpectError: regexp.MustCompile("one of `id,name` must be specified"),
			},
			{
				Config:      `data "anxcloud_vlan" "foo" { id = "" }`,
				ExpectError: regexp.MustCompile(`Either provide a non-empty "id" or "name" to query a VLAN.`),
			},

			// expected to succeed
			{
				Config: fmt.Sprintf(`data "anxcloud_vlan" "foo" { id = %q }`, vlanID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.anxcloud_vlan.foo", "status", "Active"),
				),
			},
			{
				Config: fmt.Sprintf(`data "anxcloud_vlan" "foo" { name = %q }`, vlanName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.anxcloud_vlan.foo", "status", "Active"),
				),
			},
		},
	})
}
