package ffcli

import (
	"flag"
	"strings"

	"github.com/peterbourgon/ff"
)

// Command combines a main function with a flag.FlagSet, and zero or more
// subcommands. A commandline program of the form
//
//    command [flags] (subcommand [flags] ...) [--] args
//
// can be represented as a declarative tree of commands.
type Command struct {
	Name        string
	FlagSet     *flag.FlagSet
	Options     []ff.Option // applied only to this command
	Subcommands []*Command
	Main        func(args []string) error
}

// ParseAndRun parses the commandline arguments for this command and all
// subcommands recursively, and executes the main function for the terminal
// command, if it's defined.
//
// Any options passed to ParseAndRun are provided to every visited command
// during ff.Parse. To provide options to only a specific command, declare them
// in the Options field for that command. Options declared for specific commands
// take precedence over options passed to ParseAndRun.
func (c *Command) ParseAndRun(args []string, options ...ff.Option) error {
	if err := ff.Parse(c.FlagSet, args, append(options, c.Options...)...); err != nil {
		return err
	}

	if len(c.FlagSet.Args()) > 0 {
		for _, subcommand := range c.Subcommands {
			if strings.EqualFold(c.FlagSet.Args()[0], subcommand.Name) {
				return subcommand.ParseAndRun(c.FlagSet.Args()[1:], options...)
			}
		}
	}

	if c.Main != nil {
		return c.Main(c.FlagSet.Args())
	}

	return nil
}
