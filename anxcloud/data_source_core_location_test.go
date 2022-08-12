package anxcloud

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	corev1 "go.anx.io/go-anxcloud/pkg/apis/core/v1"
)

func TestAccAnxCloudCoreLocationDataSource(t *testing.T) {
	const (
		anx04Identifier = "52b5f6b2fd3a4a7eaaedf1a7c019e9ea"
		anx04Code       = "ANX04"
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			// expected to fail
			{
				Config:      testAccAnxCloudCoreLocationDataSource("foo", corev1.Location{Identifier: "DOES-NOT-EXIST"}),
				ExpectError: regexp.MustCompile("Not Found"),
			},
			{
				Config:      testAccAnxCloudCoreLocationDataSource("foo", corev1.Location{Code: "ANXDOESNOTEXIST"}),
				ExpectError: regexp.MustCompile("Not Found"),
			},
			{
				Config:      testAccAnxCloudCoreLocationDataSource("foo", corev1.Location{Identifier: "valid identifier", Code: "valid code"}),
				ExpectError: regexp.MustCompile("only one of `code,identifier` can be specified"),
			},
			{
				Config:      testAccAnxCloudCoreLocationDataSource("foo", corev1.Location{}),
				ExpectError: regexp.MustCompile("one of `code,identifier` must be specified"),
			},

			// expected to succeed
			{
				Config: testAccAnxCloudCoreLocationDataSource("foo", corev1.Location{Identifier: anx04Identifier}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.anxcloud_core_location.foo", "code", anx04Code),
				),
			},
			{
				Config: testAccAnxCloudCoreLocationDataSource("foo", corev1.Location{Code: anx04Code}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.anxcloud_core_location.foo", "identifier", anx04Identifier),
				),
			},
		},
	})
}

func testAccAnxCloudCoreLocationDataSource(dataSourceName string, location corev1.Location) string {
	attributes := ""
	if location.Identifier != "" {
		attributes += fmt.Sprintf("identifier = %q\n", location.Identifier)
	}
	if location.Code != "" {
		attributes += fmt.Sprintf("code = %q\n", location.Code)
	}

	return fmt.Sprintf(`
	data "anxcloud_core_location" "%s" {
		%s
	}
	`, dataSourceName, attributes)
}
