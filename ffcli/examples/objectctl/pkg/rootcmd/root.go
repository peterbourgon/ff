package rootcmd

import (
	"context"
	"flag"

	"github.com/peterbourgon/ff/ffcli/examples/objectctl/pkg/objectapi"
)

// Config for the root command, including flags that
// should be available to each subcommand.
type Config struct {
	Client  *objectapi.Client
	Token   string
	Verbose bool
}

// NewConfig returns a flag set with the command's flags registered,
// and a config that will have the value of those flags after parse.
func NewConfig() (*flag.FlagSet, *Config) {
	var cfg Config
	fs := flag.NewFlagSet("objectctl", flag.ExitOnError)
	fs.StringVar(&cfg.Token, "token", "", "secret token for object API")
	fs.BoolVar(&cfg.Verbose, "v", false, "log verbose output")
	return fs, &cfg
}

// Postparse initializes the client with the value of the token.
func (c *Config) Postparse() (err error) {
	c.Client, err = objectapi.NewClient(c.Token)
	return err
}

// Exec function for this command.
func (c *Config) Exec(context.Context, []string) error {
	// The root command has no meaning, so if it gets executed,
	// display the usage text to the user instead.
	return flag.ErrHelp
}
