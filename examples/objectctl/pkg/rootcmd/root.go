package rootcmd

import (
	"io"

	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/examples/objectctl/pkg/objectapi"
)

type RootConfig struct {
	Stdout  io.Writer
	Stderr  io.Writer
	Token   string
	Verbose bool
	Client  *objectapi.Client
	FlagSet *ff.CoreFlagSet
	Command *ff.Command
}

func New(stdout, stderr io.Writer) *RootConfig {
	var cfg RootConfig
	cfg.Stdout = stdout
	cfg.Stderr = stderr
	cfg.FlagSet = ff.NewSet("objectctl")
	cfg.FlagSet.StringVar(&cfg.Token, 0, "token", "", "secret token for object API")
	cfg.FlagSet.BoolVar(&cfg.Verbose, 'v', "verbose", false, "log verbose output")
	cfg.Command = &ff.Command{
		Name:    "objectctl",
		Usage:   "objectctl [FLAGS] <SUBCOMMAND> ...",
		FlagSet: cfg.FlagSet,
	}
	return &cfg
}
