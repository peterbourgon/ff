package listcmd

import (
	"context"
	"fmt"
	"text/tabwriter"
	"time"

	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/examples/objectctl/pkg/rootcmd"
)

type Config struct {
	*rootcmd.RootConfig
	WithAccessTimes bool
	Command         *ff.Command
	FlagSet         *ff.CoreFlagSet
}

func New(parent *rootcmd.RootConfig) *Config {
	var cfg Config
	cfg.RootConfig = parent
	cfg.FlagSet = ff.NewSet("list").SetParent(parent.FlagSet)
	cfg.FlagSet.BoolVar(&cfg.WithAccessTimes, 'a', "atime", false, "include last access time of each object")
	cfg.Command = &ff.Command{
		Name:      "list",
		Usage:     "objectctl list [FLAGS]",
		ShortHelp: "list available objects",
		FlagSet:   cfg.FlagSet,
		Exec:      cfg.Exec,
	}
	cfg.RootConfig.Command.Subcommands = append(cfg.RootConfig.Command.Subcommands, cfg.Command)
	return &cfg
}

func (cfg *Config) Exec(ctx context.Context, _ []string) error {
	objects, err := cfg.Client.List(ctx)
	if err != nil {
		return err
	}

	if cfg.Verbose {
		fmt.Fprintf(cfg.Stderr, "object count: %d\n", len(objects))
	}

	if len(objects) <= 0 {
		return nil
	}

	tw := tabwriter.NewWriter(cfg.Stdout, 0, 2, 2, ' ', 0)
	if cfg.WithAccessTimes {
		fmt.Fprintf(tw, "KEY\tVALUE\tATIME\n")
	} else {
		fmt.Fprintf(tw, "KEY\tVALUE\n")
	}
	for _, object := range objects {
		if cfg.WithAccessTimes {
			fmt.Fprintf(tw, "%s\t%s\t%s\n", object.Key, object.Value, object.Access.Format(time.RFC3339))
		} else {
			fmt.Fprintf(tw, "%s\t%s\n", object.Key, object.Value)
		}
	}
	tw.Flush()

	return nil
}
