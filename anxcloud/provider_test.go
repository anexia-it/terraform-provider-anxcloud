package anxcloud

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/onsi/gomega/ghttp"
	"github.com/stretchr/testify/assert"
	"go.anx.io/go-anxcloud/pkg/api"
	"go.anx.io/go-anxcloud/pkg/client"

	tt "github.com/mitchellh/go-testing-interface"
)

var testAccProviderFactories map[string]func() (*schema.Provider, error)
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider()
	testAccProviderFactories = map[string]func() (*schema.Provider, error){
		//nolint:unparam
		// ^ signature set by tf framework; test provider does not produce any errors
		"anxcloud": func() (*schema.Provider, error) {
			return Provider(), nil
		},
	}

}

func testAccProviderFactoriesWithMockedClient(t tt.T, srv *ghttp.Server) map[string]func() (*schema.Provider, error) {
	logger := NewTerraformr(log.Default().Writer())
	opts := []client.Option{
		client.BaseURL(srv.URL()),
		client.IgnoreMissingToken(),
		client.Logger(logger.WithName("mock-client")),
		client.UserAgent(fmt.Sprintf("%s/%s (%s)", "terraform-provider-anxcloud", providerVersion, runtime.GOOS)),
	}

	c, err := api.NewAPI(api.WithClientOptions(opts...))
	if err != nil {
		assert.FailNow(t, "failed initializing mock client", err)
	}

	lc, err := client.New(opts...)
	if err != nil {
		assert.FailNow(t, "failed initializing mock client", err)
	}
	return map[string]func() (*schema.Provider, error){
		//nolint:unparam
		// ^ signature set by tf framework; test provider does not produce any errors
		"anxcloud": func() (*schema.Provider, error) {
			provider := Provider()
			provider.ConfigureContextFunc = func(ctx context.Context, rd *schema.ResourceData) (interface{}, diag.Diagnostics) {
				return providerContext{api: c, legacyClient: lc}, nil
			}
			return provider, nil
		},
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ *schema.Provider = Provider()
}

func testAccPreCheck(t *testing.T) {
	if err := os.Getenv("ANEXIA_TOKEN"); err == "" {
		t.Fatal("ANEXIA_TOKEN must be set for acceptance tests")
	}

	ctx := context.Background()
	if err := testAccProvider.Configure(ctx, terraform.NewResourceConfigRaw(nil)); err != nil && err.HasError() {
		t.Fatalf("failed to configure testAccProvider: %#v", err)
	}
}
