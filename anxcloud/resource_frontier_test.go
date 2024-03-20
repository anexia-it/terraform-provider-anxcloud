package anxcloud

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/anexia-it/terraform-provider-anxcloud/anxcloud/testutils/environment"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAnxCloudFrontier(t *testing.T) {
	environment.SkipIfNoEnvironment(t)

	runName := environment.GetEnvInfo(t).TestRunName

	apiConfig := `
	resource "anxcloud_frontier_api" "foo" {
		name = "terraform-test-%s"
		transfer_protocol = "http"
	}
	`

	endpointConfig := `
	resource "anxcloud_frontier_endpoint" "foo" {
		name = "terraform-test-%s"
		path = "bar/baz"
		api = anxcloud_frontier_api.foo.id
	}
	`

	actionConfig := `
	resource "anxcloud_frontier_action" "foo" {
		http_request_method = "get"
		endpoint = anxcloud_frontier_endpoint.foo.id

		mock_response {
			body = "%s"
			language = "plaintext"
		}
	}
	`

	deploymentConfig := `
	resource "anxcloud_frontier_deployment" "foo" {
		slug = "foo"
		api = anxcloud_frontier_api.foo.id

		revision = "%s"

		depends_on = [
			anxcloud_frontier_action.foo
		]

		lifecycle {
			create_before_destroy = true
		}
	}
	`

	checkMockEndpoint := func(expectedResponse string) resource.TestCheckFunc {
		return func(s *terraform.State) error {
			api, ok := s.RootModule().Resources["anxcloud_frontier_api.foo"]
			if !ok {
				return fmt.Errorf("anxcloud_frontier_api.foo not found in state")
			}

			// wait a few seconds for frontier to update
			time.Sleep(20 * time.Second)

			resp, err := http.Get(fmt.Sprintf("https://frontier.anexia-it.com/%s/foo/bar/baz", api.Primary.ID))
			if err != nil {
				return fmt.Errorf("http: get mock frontier endpoint: %w", err)
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("read response body: %w", err)
			}

			if string(body) != expectedResponse {
				return fmt.Errorf("unexpected response: %s", body)
			}

			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			// test create new
			{
				Config: strings.Join([]string{
					fmt.Sprintf(apiConfig, runName),
					fmt.Sprintf(endpointConfig, runName),
					fmt.Sprintf(actionConfig, "foo bar baz"),
					fmt.Sprintf(deploymentConfig, "1"),
				}, "\n"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("anxcloud_frontier_deployment.foo", "state", "deployed"),
					checkMockEndpoint("foo bar baz"),
				),
			},
			// test importability
			{
				ResourceName:      "anxcloud_frontier_api.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "anxcloud_frontier_endpoint.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "anxcloud_frontier_action.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:            "anxcloud_frontier_deployment.foo",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revision"},
			},
			// test changes
			{
				Config: strings.Join([]string{
					fmt.Sprintf(apiConfig, runName+"-changed"),
					fmt.Sprintf(endpointConfig, runName+"-changed"),
					fmt.Sprintf(actionConfig, "baz bar foo"),
					fmt.Sprintf(deploymentConfig, "2"),
				}, "\n"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("anxcloud_frontier_deployment.foo", "state", "deployed"),
					checkMockEndpoint("baz bar foo"),
				),
			},
		},
	})
}
