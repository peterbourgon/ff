package createcmd

import (
	"context"
	"errors"
	"flag"

	"github.com/peterbourgon/ff/ffcli/examples/objectctl/pkg/rootcmd"
)

// Config for the create subcommand, including a reference
// to the global config, for access to global flags.
type Config struct {
	Global    *rootcmd.Config
	Overwrite bool
}

// NewConfig returns a flag set with the command's flags registered,
// and a config that will have the value of those flags after parse.
func NewConfig(global *rootcmd.Config) (*flag.FlagSet, *Config) {
	cfg := Config{Global: global}
	fs := flag.NewFlagSet("objectctl delete", flag.ExitOnError)
	fs.BoolVar(&cfg.Overwrite, "overwrite", false, "overwrite existing object, if it exists")
	return fs, &cfg
}

// Exec function for this command.
func (c *Config) Exec(context.Context, []string) error {
	return errors.New("not implemented")
}
