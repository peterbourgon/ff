package main

import (
	"github.com/peterbourgon/ff/ffcli/examples/2-objectctl/pkg/createcmd"
	"github.com/peterbourgon/ff/ffcli/examples/2-objectctl/pkg/deletecmd"
	"github.com/peterbourgon/ff/ffcli/examples/2-objectctl/pkg/listcmd"
	"github.com/peterbourgon/ff/ffcli/examples/2-objectctl/pkg/rootcmd"
)

func main() {
	var (
		rootFlagSet, rootConfig     = rootcmd.NewConfig()
		listFlagSet, listConfig     = listcmd.NewConfig(rootConfig)
		createFlagSet, createConfig = createcmd.NewConfig(rootConfig)
		deleteFlagSet, deleteConfig = deletecmd.NewConfig(rootConfig)
	)

	// TODO(pb): elide flagsets and return commands directly
}
