package anxcloud

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/anexia-it/terraform-provider-anxcloud/anxcloud/testutils/environment"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	e5ev1 "go.anx.io/go-anxcloud/pkg/apis/e5e/v1"
)

func TestAccAnxCloudE5EApplication(t *testing.T) {
	environment.SkipIfNoEnvironment(t)

	runName := environment.GetEnvInfo(t).TestRunName

	applicationConfig := `
	resource "anxcloud_e5e_application" "foo" {
		name = "terraform-test-%s"
	}
	`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(applicationConfig, runName),
			},
			{
				ResourceName:      "anxcloud_e5e_application.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: fmt.Sprintf(applicationConfig, runName+"-changed"),
			},
		},
	})
}

func TestAccAnxCloudE5EFunction(t *testing.T) {
	environment.SkipIfNoEnvironment(t)

	runName := environment.GetEnvInfo(t).TestRunName

	applicationConfig := fmt.Sprintf(`
	resource "anxcloud_e5e_application" "foo" {
		name = "terraform-test-function-%s"
	}
	`, runName)

	testFunction := func(name string) e5ev1.Function {
		return e5ev1.Function{
			Name:           name,
			State:          "disabled",
			Runtime:        "python_310",
			Entrypoint:     "foo::Bar",
			StorageBackend: "s3",
			StorageBackendMeta: &e5ev1.StorageBackendMeta{
				StorageBackendMetaS3: &e5ev1.StorageBackendMetaS3{
					Endpoint:   "https://foo.bar",
					BucketName: "foo",
					ObjectPath: "bar",
					AccessKey:  "foo",
					SecretKey:  "bar",
				},
				StorageBackendMetaGit: &e5ev1.StorageBackendMetaGit{},
			},
			EnvironmentVariables: &[]e5ev1.EnvironmentVariable{
				{Name: "foo", Value: "bar"},
				{Name: "bar", Value: "foo"},
			},
			Hostnames: &[]e5ev1.Hostname{
				{Hostname: "foo", IP: "198.51.100.1"},
				{Hostname: "bar", IP: "198.51.100.2"},
			},
			KeepAlive:        50,
			QuotaStorage:     20,
			QuotaTimeout:     20,
			QuotaConcurrency: 30,
			QuotaMemory:      128,
			QuotaCPU:         70,
			WorkerType:       e5ev1.WorkerTypeStandard,
		}
	}

	testFunc := testFunction(fmt.Sprintf("terraform-test-function-%s", runName))

	testFuncChanged := testFunction(fmt.Sprintf("terraform-test-function-%s-changed", runName))
	testFuncChanged.StorageBackend = "git"
	testFuncChanged.StorageBackendMeta = nil
	(*testFuncChanged.EnvironmentVariables) = append((*testFuncChanged.EnvironmentVariables), e5ev1.EnvironmentVariable{
		Name:   "baz",
		Value:  "foobar",
		Secret: true,
	})

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: strings.Join([]string{
					applicationConfig,
					testAccAnxCloudE5EFunctionRenderResource(testFunc),
				}, "\n"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("anxcloud_e5e_application.foo", "id", "anxcloud_e5e_function.foo", "application"),
					testAccAnxCloudE5EFunctionExists(testFunc),
				),
			},
			{
				ResourceName:      "anxcloud_e5e_function.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: strings.Join([]string{
					applicationConfig,
					testAccAnxCloudE5EFunctionRenderResource(testFuncChanged),
				}, "\n"),
			},
		},
	})
}
func testAccAnxCloudE5EFunctionExists(expected e5ev1.Function) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rsFunc, ok := s.RootModule().Resources["anxcloud_e5e_function.foo"]
		if !ok {
			return fmt.Errorf("e5e function not found in state")
		}

		rsApp, ok := s.RootModule().Resources["anxcloud_e5e_application.foo"]
		if !ok {
			return fmt.Errorf("e5e application not found in state")
		}

		a := apiFromProviderConfig(testAccProvider.Meta())

		engineFunctionState := e5ev1.Function{Identifier: rsFunc.Primary.ID}
		if err := a.Get(context.TODO(), &engineFunctionState); err != nil {
			return fmt.Errorf("fetch function state from engine: %w", err)
		}

		expected.Identifier = rsFunc.Primary.ID
		expected.ApplicationIdentifier = rsApp.Primary.ID
		expected.DeploymentState = engineFunctionState.DeploymentState

		if diff := cmp.Diff(expected, engineFunctionState, cmpopts.IgnoreUnexported(e5ev1.Function{})); diff != "" {
			return fmt.Errorf("e5ev1.Function mismatch (-want +got):\n%s", diff)
		}

		return nil
	}
}

func testAccAnxCloudE5EFunctionRenderStorageBackend(backendType e5ev1.StorageBackend) string {
	if backendType == e5ev1.StorageBackendS3 {
		return `
			storage_backend_s3 {
				endpoint = "https://foo.bar"
				bucket_name = "foo"
				object_path = "bar"
				access_key = "foo"
				secret_key = "bar"
			}
			`
	} else if backendType == e5ev1.StorageBackendGit {
		return `
			storage_backend_git {
				url = "https://foo.bar/foo.git"
				username = "foo"
				password = "bar"
			}
			`
	}
	return ""
}

func testAccAnxCloudE5EFunctionRenderEnvironmentVariables(vars []e5ev1.EnvironmentVariable) string {
	var output strings.Builder

	for _, variable := range vars {
		output.WriteString(fmt.Sprintf(`
			env {
				name = "%s"
				value = "%s"
				secret = %t
			}
			`, variable.Name, variable.Value, variable.Secret))
	}

	return output.String()
}

func testAccAnxCloudE5EFunctionRenderHostnames(hostnames []e5ev1.Hostname) string {
	var output strings.Builder

	for _, hostname := range hostnames {
		output.WriteString(fmt.Sprintf(`
			hostname {
				hostname = "%s"
				ip = "%s"
			}
			`, hostname.Hostname, hostname.IP))
	}

	return output.String()
}

func testAccAnxCloudE5EFunctionRenderResource(config e5ev1.Function) string {
	return fmt.Sprintf(`
		resource "anxcloud_e5e_function" "foo" {
			name = "%s"
			application = anxcloud_e5e_application.foo.id
			runtime = "%s"
			entrypoint = "%s"

			# storage backend config
			%s

			# environment variables
			%s

			# hostnames
			%s

			keep_alive = %d

			quota_storage = %d
			quota_timeout = %d
			quota_concurrency = %d
			quota_memory = %d
			quota_cpu = %d

			worker_type = "%s"
		}
		`,
		config.Name,
		config.Runtime,
		config.Entrypoint,
		testAccAnxCloudE5EFunctionRenderStorageBackend(config.StorageBackend),
		testAccAnxCloudE5EFunctionRenderEnvironmentVariables(*config.EnvironmentVariables),
		testAccAnxCloudE5EFunctionRenderHostnames(*config.Hostnames),
		config.KeepAlive,
		config.QuotaStorage,
		config.QuotaTimeout,
		config.QuotaConcurrency,
		config.QuotaMemory,
		config.QuotaCPU,
		config.WorkerType,
	)
}
