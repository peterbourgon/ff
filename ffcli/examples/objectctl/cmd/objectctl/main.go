package main

import (
	"context"
	"fmt"
	"os"

	"github.com/peterbourgon/ff/ffcli"
	"github.com/peterbourgon/ff/ffcli/examples/objectctl/pkg/createcmd"
	"github.com/peterbourgon/ff/ffcli/examples/objectctl/pkg/deletecmd"
	"github.com/peterbourgon/ff/ffcli/examples/objectctl/pkg/listcmd"
	"github.com/peterbourgon/ff/ffcli/examples/objectctl/pkg/rootcmd"
)

func main() {
	var (
		out                     = os.Stdout
		rootCommand, rootConfig = rootcmd.New()
		createCommand           = createcmd.New(rootConfig, out)
		deleteCommand           = deletecmd.New(rootConfig, out)
		listCommand             = listcmd.New(rootConfig, out)
	)

	rootCommand.Subcommands = []*ffcli.Command{
		createCommand,
		deleteCommand,
		listCommand,
	}

	if err := rootCommand.Run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintf(out, "error: %v\n", err)
	}
}
