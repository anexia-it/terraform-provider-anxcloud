package main

import (
	"flag"

	"github.com/anexia-it/terraform-provider-anxcloud/anxcloud"

	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main() {
	var debugMode bool

	flag.BoolVar(&debugMode, "debuggable", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := &plugin.ServeOpts{
		Debug:        debugMode,
		ProviderAddr: "registry.terraform.io/hashicorp/anxcloud",
		ProviderFunc: anxcloud.Provider,
	}

	plugin.Serve(opts)
}
