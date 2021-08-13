package main

import (
	"context"
	"flag"
	"log"

	"github.com/anexia-it/terraform-provider-anxcloud/anxcloud"

	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main() {
	var debugMode bool

	flag.BoolVar(&debugMode, "debuggable", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	if debugMode {
		err := plugin.Debug(context.Background(), "hashicorp.com/anexia-it/anxcloud",
			&plugin.ServeOpts{
				ProviderFunc: anxcloud.Provider,
			})
		if err != nil {
			log.Println(err.Error())
		}
	} else {
		plugin.Serve(&plugin.ServeOpts{
			ProviderFunc: anxcloud.Provider})
	}
}
