package rootcmd

import (
	"io"

	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/examples/objectctl/pkg/objectapi"
	"github.com/peterbourgon/ff/v4/ffhelp"
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
		LongHelp: ffhelp.Rewrap(`
			Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nam diam eros,
			vestibulum at pulvinar vulputate, vehicula id lacus. Class aptent taciti
			sociosqu ad litora torquent per conubia nostra, per inceptos himenaeos. Mauris
			venenatis felis orci, ac consectetur mi molestie ac. Integer pharetra pharetra
			odio. Maecenas metus eros, viverra eget efficitur ut, feugiat in tortor.
		`),
		Usage: "objectctl [FLAGS] <SUBCOMMAND> ...",
		Flags: cfg.Flags,
	}
	return &cfg
}
