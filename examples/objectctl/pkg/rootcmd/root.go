package rootcmd

import (
	"io"

	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/examples/objectctl/pkg/objectapi"
	"github.com/peterbourgon/ff/v4/ffval"
)

type RootConfig struct {
	Stdout  io.Writer
	Stderr  io.Writer
	Token   string
	Verbose bool
	Client  *objectapi.Client
	Flags   *ff.FlagSet
	Command *ff.Command
}

func New(stdout, stderr io.Writer) *RootConfig {
	var cfg RootConfig
	cfg.Stdout = stdout
	cfg.Stderr = stderr
	cfg.Flags = ff.NewFlagSet("objectctl")
	cfg.Flags.StringVar(&cfg.Token, 0, "token", "", "secret token for object API")
	cfg.Flags.AddFlag(ff.FlagConfig{
		ShortName: 'v',
		LongName:  "verbose",
		Value:     ffval.NewValue(&cfg.Verbose),
		Usage:     "log verbose output",
		NoDefault: true,
	})
	cfg.Command = &ff.Command{
		Name:      "objectctl",
		ShortHelp: "control objects",
		Usage:     "objectctl [FLAGS] <SUBCOMMAND> ...",
		Flags:     cfg.Flags,
	}
	return &cfg
}
