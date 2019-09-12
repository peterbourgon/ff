package deletecmd

import (
	"context"
	"errors"
	"flag"
	"io"

	"github.com/peterbourgon/ff/ffcli"
	"github.com/peterbourgon/ff/ffcli/examples/objectctl/pkg/rootcmd"
)

// Deleter models the Delete method of an objectapi.Client.
type Deleter interface {
	Delete(key string) (deleted bool, err error)
}

// Config for the delete subcommand, including a reference to the API client.
type Config struct {
	rootConfig *rootcmd.Config
	out        io.Writer
	force      bool
}

// New TODO
func New(rootConfig *rootcmd.Config, out io.Writer) *ffcli.Command {
	cfg := Config{
		rootConfig: rootConfig,
		out:        out,
	}

	fs := flag.NewFlagSet("objectctl delete", flag.ExitOnError)
	fs.BoolVar(&cfg.force, "f", false, "force delete without confirmation")

	return &ffcli.Command{
		Name:      "delete",
		Usage:     "objectctl delete [flags] <key>",
		ShortHelp: "Delete an object",
		FlagSet:   fs,
		Exec:      cfg.Exec,
	}
}

// Exec function for this command.
func (c *Config) Exec(context.Context, []string) error {
	return errors.New("not implemented")
}
