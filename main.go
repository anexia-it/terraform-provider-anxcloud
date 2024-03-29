package main

import (
	"context"
	"flag"
	"log"

	"github.com/anexia-it/terraform-provider-anxcloud/anxcloud"
	"github.com/anexia-it/terraform-provider-anxcloud/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/tf6server"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"
)

var version = "development"

func main() {
	ctx := context.Background()

	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	upgradedSdkServer, err := tf5to6server.UpgradeServer(
		ctx,
		anxcloud.Provider(version).GRPCProvider,
	)
	if err != nil {
		log.Fatal(err)
	}

	providers := []func() tfprotov6.ProviderServer{
		providerserver.NewProtocol6(provider.New(version)()),
		func() tfprotov6.ProviderServer {
			return upgradedSdkServer
		},
	}

	muxServer, err := tf6muxserver.NewMuxServer(ctx, providers...)

	if err != nil {
		log.Fatal(err)
	}

	var serveOpts []tf6server.ServeOpt

	if debug {
		serveOpts = append(serveOpts, tf6server.WithManagedDebug())
	}

	err = tf6server.Serve(
		"registry.terraform.io/anexia-it/anxcloud",
		muxServer.ProviderServer,
		serveOpts...,
	)

	if err != nil {
		log.Fatal(err)
	}
}
