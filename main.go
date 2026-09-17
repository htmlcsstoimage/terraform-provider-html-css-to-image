package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/htmlcsstoimage/terraform-provider-html-css-to-image/internal/provider"
)

var version = "dev"

func main() {
	debug := flag.Bool("debug", false, "Enable debugger support")
	flag.Parse()
	if err := providerserver.Serve(context.Background(), provider.New(version), providerserver.ServeOpts{Address: "registry.terraform.io/htmlcsstoimage/html-css-to-image", Debug: *debug}); err != nil {
		log.Fatal(err)
	}
}
