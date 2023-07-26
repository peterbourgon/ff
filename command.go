package ff

import (
	"context"
	"fmt"
	"strings"
)

// Command is a declarative structure that combines a main function with a flag
// set and zero or more subcommands. It's intended to model CLI applications
// which can be be represented as a tree of such commands.
type Command struct {
	// Name of the command, which is used when producing the help output for the
	// command, as well as for subcommand matching.
	//
	// Required.
	Name string

	// Usage is a single line which should describe the syntax of the command,
	// including flags and arguments. It's typically printed at the top of the
	// help output for the command. For example,
	//
	//    USAGE
	//      cmd [FLAGS] subcmd [FLAGS] <ARG> [<ARG>...]
	//
	// Here, the usage string begins with "cmd [FLAGS] ...".
	//
	// Recommended. If not provided, the help output for the command should not
	// include a usage section.
	Usage string

	// ShortHelp is a single line which should very briefly describe the purpose
	// of the command in prose. It's typically printed next to the command name
	// when it appears as a subcommand in help output. For example,
	//
	//    SUBCOMMANDS
	//      commandname   this is the short help string
	//
	// Recommended.
	ShortHelp string

	// LongHelp is a multi-line string, usually one or more paragraphs of prose,
	// which explain the command in detail. It's typically included in the help
	// output for the command, separate from other sections.
	//
	// Long help should be formatted for user readability. For example, if help
	// output is written to a terminal, long help should hard-wrap lines at an
	// appropriate column width for that terminal.
	//
	// Optional.
	LongHelp string

	// FlagSet is the set of flags associated with, and parsed by, this command.
	//
	// When building a command tree, it's often useful to allow flags defined by
	// parent commands to be specified by any subcommand. The core flag set
	// supports this behavior via SetParent, see the documentation of that
	// method for details.
	//
	// Optional. If not provided, an empty flag set will be constructed and used
	// so that the -h, --help flag works as expected.
	FlagSet FlagSet

	// Subcommands which are available underneath (i.e. after) this command.
	// Selecting a subcommand is done via a case-insensitive comparision of the
	// first post-parse argument to this command, against the name of each
	// subcommand.
	//
	// Optional.
	Subcommands []*Command

	isParsed bool
	selected *Command
	parent   *Command
	args     []string

	// Exec is invoked by Run (or ParseAndRun) if this command was selected as
	// the terminal command during the parse phase. The args passed to Exec are
	// the args left over after parsing.
	//
	// Optional. If not provided, and this command is identified as the terminal
	// command during the parse phase, the run phase will return NoExecError.
	Exec func(context.Context, []string) error
}

// Parse the args and options against the defined command, which sets relevant
// flags, traverses the command hierarchy to select a terminal command, and
// captures the arguments that will be given to that command's exec function.
// Args should not include the name of the program: os.Args[1:], not os.Args.
func (cmd *Command) Parse(args []string, options ...Option) error {
	// Initial validation and safety checks.
	if cmd.Name == "" {
		return fmt.Errorf("name is required")
	}
	if cmd.isParsed {
		return fmt.Errorf("%s: %w", cmd.Name, ErrAlreadyParsed)
	}

	// If no flag set was given, set an empty default, so -h, --help works.
	if cmd.FlagSet == nil {
		cmd.FlagSet = NewSet(cmd.Name)
	}

	// Parse this command's flag set from the provided args.
	if err := parseFlagSet(cmd.FlagSet, args, options...); err != nil {
		cmd.selected = cmd // allow GetSelected to work even with errors
		return fmt.Errorf("%s: %w", cmd.Name, err)
	}

	// If the parse was successful, mark the command as parsed.
	cmd.isParsed = true

	// Set this command's args to the args left over after parsing.
	cmd.args = cmd.FlagSet.GetArgs()

	// If there were any args, we might need to descend to a subcommand.
	if len(cmd.args) > 0 {
		first := cmd.args[0]
		for _, subcommand := range cmd.Subcommands {
			if strings.EqualFold(first, subcommand.Name) {
				cmd.selected = subcommand
				subcommand.parent = cmd
				return subcommand.Parse(cmd.args[1:], options...)
			}
		}
	}

	// We didn't find a matching subcommand, so we selected ourselves.
	cmd.selected = cmd

	// Parse complete.
	return nil
}

// Run the exec function of the command selected during the parse phase, passing
// the args left over after parsing. Calling run without first calling parse
// will result in an error.
func (cmd *Command) Run(ctx context.Context) error {
	switch {
	case !cmd.isParsed:
		return ErrNotParsed
	case cmd.isParsed && cmd.selected == nil:
		return ErrNotParsed
	case cmd.isParsed && cmd.selected == cmd && cmd.Exec == nil:
		return fmt.Errorf("%s: %w", cmd.Name, ErrNoExec)
	case cmd.isParsed && cmd.selected == cmd && cmd.Exec != nil:
		return cmd.Exec(ctx, cmd.args)
	default:
		return cmd.selected.Run(ctx)
	}
}

// ParseAndRun calls parse and then, on success, run.
func (cmd *Command) ParseAndRun(ctx context.Context, args []string, options ...Option) error {
	if err := cmd.Parse(args, options...); err != nil {
		return fmt.Errorf("parse: %w", err)
	}

	if err := cmd.Run(ctx); err != nil {
		return fmt.Errorf("run: %w", err)
	}

	return nil
}

// GetSelected returns the terminal command selected during the parse phase, or
// nil if the command hasn't been parsed.
func (cmd *Command) GetSelected() *Command {
	if cmd.selected == nil {
		return nil
	}

	if cmd.selected == cmd {
		return cmd
	}

	return cmd.selected.GetSelected()
}

// GetParent returns the parent command of this command, or nil if a parent
// hasn't been set. Parents are set during the parse phase, but only for
// commands which are traversed.
func (cmd *Command) GetParent() *Command {
	return cmd.parent
}

// Reset every command in the command tree, including all flag sets, to their
// initial state. Flag sets must implement [Resetter], or else reset will return
// an error. After a successful reset, the command can be parsed and run as if
// it were newly constructed.
func (cmd *Command) Reset() error {
	var check func(*Command) error

	check = func(c *Command) error {
		if c.FlagSet != nil {
			if _, ok := c.FlagSet.(Resetter); !ok {
				return fmt.Errorf("flag set (%T) doesn't implement Resetter", c.FlagSet)
			}
		}
		for _, sc := range c.Subcommands {
			if err := check(sc); err != nil {
				return err
			}
		}
		return nil
	}

	if err := check(cmd); err != nil {
		return err
	}

	if cmd.FlagSet != nil {
		r, ok := cmd.FlagSet.(Resetter)
		if !ok {
			panic(fmt.Errorf("flag set (%T) doesn't implement Resetter, even after check (programmer error)", cmd.FlagSet))
		}
		if err := r.Reset(); err != nil {
			return fmt.Errorf("reset flags: %w", err)
		}
	}

	for _, subcommand := range cmd.Subcommands {
		if err := subcommand.Reset(); err != nil {
			return fmt.Errorf("%s: %w", subcommand.Name, err)
		}
	}

	cmd.isParsed = false
	cmd.selected = nil
	cmd.parent = nil
	cmd.args = []string{}

	return nil
}
