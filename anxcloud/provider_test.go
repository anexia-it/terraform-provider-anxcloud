package anxcloud

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var testAccProviderFactories map[string]func() (*schema.Provider, error)
var testAccProvider *schema.Provider

var providerVersion = "development"

func init() {
	testAccProvider = Provider(providerVersion)
	testAccProviderFactories = map[string]func() (*schema.Provider, error){
		//nolint:unparam
		// ^ signature set by tf framework; test provider does not produce any errors
		"anxcloud": func() (*schema.Provider, error) {
			return Provider(providerVersion), nil
		},
	}

}

func TestProvider(t *testing.T) {
	if err := Provider(providerVersion).InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ *schema.Provider = Provider(providerVersion)
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
