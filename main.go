package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/vpsfreecz/terraform-provider-vpsadmin/vpsadmin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{ProviderFunc: vpsadmin.Provider})
}
