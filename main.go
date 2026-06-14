package main

import (
	"context"
	"flag"
	"log"

	"github.com/SigNoz/terraform-provider-signoz/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

var (
	// these will be set by the goreleaser configuration
	// to appropriate values for the compiled binary.
	version string = "dev"

	// goreleaser can pass other information to the main package, such as the specific commit
	// https://goreleaser.com/cookbooks/using-main.version/
)

const (
	registry       = "registry.terraform.io/signoz/signoz"
	terraformAgent = "terraform"
	name           = "signoz"
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: registry,
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), provider.New(terraformAgent, version, name), opts)
	if err != nil {
		log.Fatal(err.Error())
	}
}
