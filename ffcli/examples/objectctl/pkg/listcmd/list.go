package listcmd

import (
	"context"
	"flag"
	"fmt"
	"io"
	"text/tabwriter"
	"time"

	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/peterbourgon/ff/v3/ffcli/examples/objectctl/pkg/rootcmd"
)

// Config for the list subcommand, including a reference
// to the global config, for access to global flags.
type Config struct {
	rootConfig      *rootcmd.Config
	out             io.Writer
	withAccessTimes bool
}

// New creates a new ffcli.Command for the list subcommand.
func New(rootConfig *rootcmd.Config, out io.Writer) *ffcli.Command {
	cfg := Config{
		rootConfig: rootConfig,
		out:        out,
	}

	fs := flag.NewFlagSet("objectctl list", flag.ExitOnError)
	fs.BoolVar(&cfg.withAccessTimes, "a", false, "include last access time of each object")
	rootConfig.RegisterFlags(fs)

	return &ffcli.Command{
		Name:       "list",
		ShortUsage: "objectctl list [flags] [<prefix>]",
		ShortHelp:  "List available objects",
		FlagSet:    fs,
		Exec:       cfg.Exec,
	}
}

// Exec function for this command.
func (c *Config) Exec(ctx context.Context, _ []string) error {
	objects, err := c.rootConfig.Client.List(ctx)
	if err != nil {
		return fmt.Errorf("error executing list: %w", err)
	}

	if len(objects) <= 0 {
		fmt.Fprintf(c.out, "no objects\n")
		return nil
	}

	if c.rootConfig.Verbose {
		fmt.Fprintf(c.out, "object count: %d\n", len(objects))
	}

	tw := tabwriter.NewWriter(c.out, 0, 2, 2, ' ', 0)
	if c.withAccessTimes {
		fmt.Fprintf(tw, "KEY\tVALUE\tATIME\n")
	} else {
		fmt.Fprintf(tw, "KEY\tVALUE\n")
	}
	for _, object := range objects {
		if c.withAccessTimes {
			fmt.Fprintf(tw, "%s\t%s\t%s\n", object.Key, object.Value, object.Access.Format(time.RFC3339))
		} else {
			fmt.Fprintf(tw, "%s\t%s\n", object.Key, object.Value)
		}
	}
	tw.Flush()

	return nil
}
