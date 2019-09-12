package createcmd

import (
	"context"
	"errors"
	"flag"
	"io"

	"github.com/peterbourgon/ff/ffcli"
	"github.com/peterbourgon/ff/ffcli/examples/objectctl/pkg/rootcmd"
)

// Creater models the Create method of an objectapi.Client.
type Creater interface {
	Create(key, value string, overwrite bool) error
}

// Config for the create subcommand, including a reference to the API client.
type Config struct {
	rootConfig *rootcmd.Config
	out        io.Writer
	overwrite  bool
}

// New TODO
func New(rootConfig *rootcmd.Config, out io.Writer) *ffcli.Command {
	cfg := Config{
		rootConfig: rootConfig,
		out:        out,
	}

	fs := flag.NewFlagSet("objectctl delete", flag.ExitOnError)
	fs.BoolVar(&cfg.overwrite, "overwrite", false, "overwrite existing object, if it exists")

	return &ffcli.Command{
		Name:      "create",
		Usage:     "objectctl create [flags] <key> <value data...>",
		ShortHelp: "Create or overwrite an object",
		FlagSet:   fs,
		Exec:      cfg.Exec,
	}
}

// Exec function for this command.
func (c *Config) Exec(context.Context, []string) error {
	return errors.New("not implemented")
}
