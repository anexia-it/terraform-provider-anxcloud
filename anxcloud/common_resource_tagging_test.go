package anxcloud

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	corev1 "go.anx.io/go-anxcloud/pkg/apis/core/v1"
)

func testAccAnxCloudCheckResourceTagged(resourcePath string, expectedTags ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		a := testAccProvider.Meta().(providerContext).api
		rs, ok := s.RootModule().Resources[resourcePath]
		if !ok {
			return fmt.Errorf("resource %q not found", resourcePath)
		}

		remoteTags, err := readTags(context.TODO(), a, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("failed to fetch remote tags: %w", err)
		}

		sort.Strings(expectedTags)
		sort.Strings(remoteTags)

		if !reflect.DeepEqual(expectedTags, remoteTags) {
			return fmt.Errorf("resource %s tags didn't match remote tags, got %v - expected %v", resourcePath, remoteTags, expectedTags)
		}

		return nil
	}
}

func testAccAnxCloudAddRemoteTag(resourcePath, tag string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		a := testAccProvider.Meta().(providerContext).api
		rs, ok := s.RootModule().Resources[resourcePath]
		if !ok {
			return "", fmt.Errorf("resource %q not found", resourcePath)
		}

		if err := corev1.Tag(context.TODO(), a, &corev1.Resource{Identifier: rs.Primary.ID}, tag); err != nil {
			return "", fmt.Errorf("failed to tag resource: %w", err)
		}

		return rs.Primary.ID, nil
	}
}

func testAccAnxCloudCommonResourceTagTestSteps(tpl, resourcePath string) []resource.TestStep {
	return []resource.TestStep{
		// create resource with tags
		{
			Config: fmt.Sprintf(tpl, generateTagsString("foo", "bar")),
			Check:  testAccAnxCloudCheckResourceTagged(resourcePath, "foo", "bar"),
		},
		// remove tag
		{
			Config: fmt.Sprintf(tpl, generateTagsString("foo")),
			Check:  testAccAnxCloudCheckResourceTagged(resourcePath, "foo"),
		},
		// add tag
		{
			Config: fmt.Sprintf(tpl, generateTagsString("foo", "bar", "baz")),
			Check:  testAccAnxCloudCheckResourceTagged(resourcePath, "foo", "bar", "baz"),
		},
		// change remote tags
		{
			// this should technically be a PreConfig
			// since PreConfig does not expose the *terraform.State we use this as a workaround
			ImportStateIdFunc: testAccAnxCloudAddRemoteTag(resourcePath, "foobaz"),
			ImportState:       true,
			ResourceName:      resourcePath,
		},
		// reconcile tags (should remove previously created "foobaz" tag)
		{
			Config: fmt.Sprintf(tpl, generateTagsString("foo", "bar", "baz")),
			Check:  testAccAnxCloudCheckResourceTagged(resourcePath, "foo", "bar", "baz"),
		},
		// removed tags argument -> expect remote to be untouched
		{
			Config: fmt.Sprintf(tpl, ""),
			Check:  testAccAnxCloudCheckResourceTagged(resourcePath, "foo", "bar", "baz"),
		},
	}
}

func generateTagsString(tags ...string) string {
	if len(tags) == 0 {
		return ""
	}

	ret := strings.Builder{}
	ret.WriteString("tags = [\n")

	for _, tag := range tags {
		ret.WriteString(fmt.Sprintf("%q,\n", tag))
	}

	ret.WriteString("]\n")
	return ret.String()
}

func withoutTags(tpl string) string {
	return fmt.Sprintf(tpl, "")
}
