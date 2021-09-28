package createcmd

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/peterbourgon/ff/v3/ffcli/examples/objectctl/pkg/rootcmd"
)

// Config for the create subcommand, including a reference to the API client.
type Config struct {
	rootConfig *rootcmd.Config
	out        io.Writer
	overwrite  bool
}

// New returns a usable ffcli.Command for the create subcommand.
func New(rootConfig *rootcmd.Config, out io.Writer) *ffcli.Command {
	cfg := Config{
		rootConfig: rootConfig,
		out:        out,
	}

	fs := flag.NewFlagSet("objectctl create", flag.ExitOnError)
	fs.BoolVar(&cfg.overwrite, "overwrite", false, "overwrite existing object, if it exists")
	rootConfig.RegisterFlags(fs)

	return &ffcli.Command{
		Name:       "create",
		ShortUsage: "objectctl create [flags] <key> <value data...>",
		ShortHelp:  "Create or overwrite an object",
		FlagSet:    fs,
		Exec:       cfg.Exec,
	}
}

// Exec function for this command.
func (c *Config) Exec(ctx context.Context, args []string) error {
	if len(args) < 2 {
		return errors.New("create requires at least 2 args")
	}

	var (
		key   = args[0]
		value = strings.Join(args[1:], " ")
		err   = c.rootConfig.Client.Create(ctx, key, value, c.overwrite)
	)
	if err != nil {
		return err
	}

	if c.rootConfig.Verbose {
		fmt.Fprintf(c.out, "create %q OK\n", key)
	}

	return nil
}
