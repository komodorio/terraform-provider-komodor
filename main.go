package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/komodorio/terraform-provider-komodor/komodor"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: komodor.Provider})
}
