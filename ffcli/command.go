package ffcli

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

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
	Usage       string
	ShortHelp   string
	LongHelp    string
	FlagSet     *flag.FlagSet
	Options     []ff.Option // applied only to this command
	Subcommands []*Command
	Exec        func(args []string) error
}

// Run parses the commandline arguments for this command and all subcommands
// recursively, and invokes the Exec function for the terminal command, if it's
// defined.
//
// Any options passed to Run are provided to every visited command during
// ff.Parse. To provide options to only a specific command, declare them in the
// Options field for that command. Options declared for specific commands take
// precedence over options passed to Run.
func (c *Command) Run(args []string, options ...ff.Option) error {
	if c.FlagSet == nil {
		c.FlagSet = flag.NewFlagSet(c.Name, flag.ExitOnError) // TODO(pb)
	}

	c.FlagSet.Usage = c.usage
	if err := ff.Parse(c.FlagSet, args, append(options, c.Options...)...); err != nil {
		return err
	}

	if len(c.FlagSet.Args()) > 0 {
		for _, subcommand := range c.Subcommands {
			if strings.EqualFold(c.FlagSet.Args()[0], subcommand.Name) {
				return subcommand.Run(c.FlagSet.Args()[1:], options...)
			}
		}
	}

	if c.Exec != nil {
		return c.Exec(c.FlagSet.Args())
	}

	return nil
}

func (c *Command) usage() {
	fmt.Fprintf(os.Stdout, "USAGE\n")
	if c.Usage != "" {
		fmt.Fprintf(os.Stdout, "  %s\n", c.Usage)
	} else {
		fmt.Fprintf(os.Stdout, "  %s\n", c.Name)
	}
	fmt.Fprintf(os.Stdout, "\n")

	if c.LongHelp != "" {
		fmt.Fprintf(os.Stdout, "%s\n\n", c.LongHelp)
	}

	if len(c.Subcommands) > 0 {
		fmt.Fprintf(os.Stdout, "SUBCOMMANDS\n")
		tw := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
		for _, subcommand := range c.Subcommands {
			fmt.Fprintf(tw, "  %s\t%s\n", subcommand.Name, subcommand.ShortHelp)
		}
		tw.Flush()
		fmt.Fprintf(os.Stdout, "\n")
	}

	if countFlags(c.FlagSet) > 0 {
		fmt.Fprintf(os.Stdout, "FLAGS\n")
		tw := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
		c.FlagSet.VisitAll(func(f *flag.Flag) {
			def := f.DefValue
			if def == "" {
				def = "..."
			}
			fmt.Fprintf(tw, "  -%s %s\t%s\n", f.Name, def, f.Usage)
		})
		tw.Flush()
		fmt.Fprintf(os.Stdout, "\n")
	}
}

func countFlags(fs *flag.FlagSet) (n int) {
	fs.VisitAll(func(*flag.Flag) { n++ })
	return n
}
