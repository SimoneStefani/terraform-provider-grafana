package main

import (
	"github.com/SimoneStefani/terraform-provider-grafana/grafana"
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{ProviderFunc: grafana.Provider})
}
