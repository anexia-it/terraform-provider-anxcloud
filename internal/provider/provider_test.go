package provider

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"testing"

	"github.com/anexia-it/terraform-provider-anxcloud/anxcloud"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"
	"go.anx.io/go-anxcloud/pkg/client"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"anxcloud": func() (tfprotov6.ProviderServer, error) {
		ctx := context.TODO()
		upgradedSdkServer, err := tf5to6server.UpgradeServer(
			ctx,
			anxcloud.Provider("test").GRPCProvider,
		)
		if err != nil {
			return nil, err
		}

		providers := []func() tfprotov6.ProviderServer{
			providerserver.NewProtocol6(New("test")()),
			func() tfprotov6.ProviderServer {
				return upgradedSdkServer
			},
		}

		muxServer, err := tf6muxserver.NewMuxServer(ctx, providers...)

		if err != nil {
			return nil, err
		}

		return muxServer.ProviderServer(), nil
	},
}

//nolint:unused
var testAccProtoV6MockProviderFactories = func(endpoint string) map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"anxcloud": func() (tfprotov6.ProviderServer, error) {
			return providerserver.NewProtocol6WithError(NewAnexiaMockProvider(endpoint))()
		},
	}
}

type anexiaMockProvider struct {
	AnexiaProvider
	endpoint string
}

func NewAnexiaMockProvider(endpoint string) provider.Provider {
	return &anexiaMockProvider{
		endpoint: endpoint,
	}
}

func (p *anexiaMockProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	logger := anxcloud.NewTerraformr(log.Default().Writer())
	opts := []client.Option{
		client.BaseURL(p.endpoint),
		client.IgnoreMissingToken(),
		client.Logger(logger.WithName("client")),
		client.UserAgent(fmt.Sprintf("%s/%s (%s)", "terraform-provider-anxcloud", p.version, runtime.GOOS)),
	}

	resp.ResourceData = opts
	resp.DataSourceData = opts
}

func testAccPreCheck(t *testing.T) {}

func TestFrameworkSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "framework suite")
}
