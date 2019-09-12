package rootcmd

import (
	"context"
	"flag"

	"github.com/peterbourgon/ff/ffcli"
	"github.com/peterbourgon/ff/ffcli/examples/objectctl/pkg/objectapi"
)

// Config for the root command, including flags and types that should be
// available to each subcommand.
type Config struct {
	token   string
	Verbose bool
	Client  *objectapi.Client
}

// New TODO
func New() (*ffcli.Command, *Config) {
	var cfg Config

	fs := flag.NewFlagSet("objectctl", flag.ExitOnError)
	fs.StringVar(&cfg.token, "token", "", "secret token for object API")
	fs.BoolVar(&cfg.Verbose, "v", false, "log verbose output")

	return &ffcli.Command{
		Name:      "objectctl",
		Usage:     "objectctl [flags] <subcommand> [flags] [<arg>...]",
		FlagSet:   fs,
		Postparse: cfg.Postparse,
		Exec:      cfg.Exec,
	}, &cfg
}

// Postparse initializes the client with the value of the token.
func (c *Config) Postparse(ctx context.Context) (err error) {
	c.Client, err = objectapi.NewClient(c.token)
	return err
}

// Exec function for this command.
func (c *Config) Exec(context.Context, []string) error {
	// The root command has no meaning, so if it gets executed,
	// display the usage text to the user instead.
	return flag.ErrHelp
}
