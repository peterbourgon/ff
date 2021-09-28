package rootcmd

import (
	"context"
	"flag"

	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/peterbourgon/ff/v3/ffcli/examples/objectctl/pkg/objectapi"
)

// Config for the root command, including flags and types that should be
// available to each subcommand.
type Config struct {
	Token   string
	Verbose bool

	Client *objectapi.Client
}

// New constructs a usable ffcli.Command and an empty Config. The config's token
// and verbose fields will be set after a successful parse. The caller must
// initialize the config's object API client field.
func New() (*ffcli.Command, *Config) {
	var cfg Config

	fs := flag.NewFlagSet("objectctl", flag.ExitOnError)
	cfg.RegisterFlags(fs)

	return &ffcli.Command{
		Name:       "objectctl",
		ShortUsage: "objectctl [flags] <subcommand> [flags] [<arg>...]",
		FlagSet:    fs,
		Exec:       cfg.Exec,
	}, &cfg
}

// RegisterFlags registers the flag fields into the provided flag.FlagSet. This
// helper function allows subcommands to register the root flags into their
// flagsets, creating "global" flags that can be passed after any subcommand at
// the commandline.
func (c *Config) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&c.Token, "token", "", "secret token for object API")
	fs.BoolVar(&c.Verbose, "v", false, "log verbose output")
}

// Exec function for this command.
func (c *Config) Exec(context.Context, []string) error {
	// The root command has no meaning, so if it gets executed,
	// display the usage text to the user instead.
	return flag.ErrHelp
}
