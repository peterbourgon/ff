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
	Flags   *ff.CoreFlags
	Command *ff.Command
}

func New(stdout, stderr io.Writer) *RootConfig {
	var cfg RootConfig
	cfg.Stdout = stdout
	cfg.Stderr = stderr
	cfg.Flags = ff.NewFlags("objectctl")
	cfg.Flags.StringVar(&cfg.Token, 0, "token", "", "secret token for object API")
	cfg.Flags.BoolVar(&cfg.Verbose, 'v', "verbose", false, "log verbose output")
	cfg.Command = &ff.Command{
		Name:  "objectctl",
		Usage: "objectctl [FLAGS] <SUBCOMMAND> ...",
		Flags: cfg.Flags,
	}
	return &cfg
}
