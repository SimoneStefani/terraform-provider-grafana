package main

import (
	"github.com/SimoneStefani/terraform-provider-grafana/grafana"
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: func() terraform.ResourceProvider {
			return grafana.Provider()
		},
	})
}
