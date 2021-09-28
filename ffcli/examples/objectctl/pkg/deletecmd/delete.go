package deletecmd

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/peterbourgon/ff/v3/ffcli/examples/objectctl/pkg/rootcmd"
)

// Config for the delete subcommand, including a reference to the API client.
type Config struct {
	rootConfig *rootcmd.Config
	out        io.Writer
	force      bool
}

// New returns a usable ffcli.Command for the delete subcommand.
func New(rootConfig *rootcmd.Config, out io.Writer) *ffcli.Command {
	cfg := Config{
		rootConfig: rootConfig,
		out:        out,
	}

	fs := flag.NewFlagSet("objectctl delete", flag.ExitOnError)
	rootConfig.RegisterFlags(fs)

	return &ffcli.Command{
		Name:       "delete",
		ShortUsage: "objectctl delete <key>",
		ShortHelp:  "Delete an object",
		FlagSet:    fs,
		Exec:       cfg.Exec,
	}
}

// Exec function for this command.
func (c *Config) Exec(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return errors.New("delete requires at least 1 arg")
	}

	var (
		key          = args[0]
		existed, err = c.rootConfig.Client.Delete(ctx, key, c.force)
	)
	if err != nil {
		return err
	}

	if c.rootConfig.Verbose {
		fmt.Fprintf(c.out, "delete %q OK (existed %v)\n", key, existed)
	}

	return nil
}
