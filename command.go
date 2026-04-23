package ff

import (
	"context"
	"fmt"
	"strings"
)

// Command is a declarative structure that combines a main function with a flag
// set and zero or more subcommands. It's intended to model CLI applications
// which can be represented as a tree of such commands.
type Command struct {
	// Name of the command, which is used when producing the help text for the
	// command, as well as for subcommand matching.
	//
	// Required.
	Name string

	// Usage is a single line string which should describe the syntax of the
	// command, including flags and arguments. It's typically printed at the top
	// of the help text for the command. For example,
	//
	//    USAGE
	//      cmd [FLAGS] subcmd [FLAGS] <ARG> [<ARG>...]
	//
	// Here, the usage string begins with "cmd [FLAGS] ...".
	//
	// Recommended. If not provided, the help text for the command should not
	// include a usage section.
	Usage string

	// ShortHelp is a single line which should very briefly describe the purpose
	// of the command in prose. It's typically printed next to the command name
	// when it appears as a subcommand in help text. For example,
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
	// output is written to a terminal, long help should include newlines which
	// hard-wrap the string at an appropriate column width for that terminal.
	//
	// Optional.
	LongHelp string

	// Flags is the set of flags associated with, and parsed by, this command.
	//
	// When building a command tree, it's often useful to allow flags defined by
	// parent commands to be specified by any subcommand. A FlagSet supports
	// this behavior via SetParent, see the documentation of that method for
	// details.
	//
	// Optional. If not provided, an empty flag set will be constructed and used
	// so that the -h, --help flag works as expected.
	Flags *FlagSet

	// Subcommands which are available underneath (i.e. after) this command.
	// Selecting a subcommand is done via a case-insensitive comparison of the
	// first post-parse argument to this command, against the name of each
	// subcommand.
	//
	// Optional.
	Subcommands []*Command

	isParsed bool
	selected *Command
	parent   *Command

	// Exec is invoked by Run (or ParseAndRun) if this command was selected as
	// the terminal command during the parse phase. The args passed to Exec are
	// the args left over after parsing.
	//
	// Optional. If not provided, running this command will result in ErrNoExec.
	Exec func(ctx context.Context, args []string) error
}

// Parse the args and options against the defined command, which sets relevant
// flags, traverses the command hierarchy to select a terminal command, and
// captures the arguments that will be given to that command's exec function.
// The args should not include the program name: pass os.Args[1:], not os.Args.
func (cmd *Command) Parse(args []string, options ...Option) error {
	// Initial validation and safety checks.
	if cmd.Name == "" {
		return fmt.Errorf("name is required")
	}
	if cmd.isParsed {
		return fmt.Errorf("%s: %w", cmd.Name, ErrAlreadyParsed)
	}

	// If no flag set was given, set an empty default, so -h, --help works.
	if cmd.Flags == nil {
		cmd.Flags = NewFlagSet(cmd.Name)
	}

	// Parse this command's flag set from the provided args.
	if err := parse(cmd.Flags, args, options...); err != nil {
		cmd.selected = cmd // allow GetSelected to work even with errors
		return fmt.Errorf("%s: %w", cmd.Name, err)
	}

	// If the parse was successful, mark the command as parsed.
	cmd.isParsed = true

	// Set this command's args to the args left over after parsing.
	commandArgs := cmd.Flags.GetArgs()

	// If there were any args, we might need to descend to a subcommand.
	if len(commandArgs) > 0 {
		first := commandArgs[0]
		for _, subcommand := range cmd.Subcommands {
			if strings.EqualFold(first, subcommand.Name) {
				cmd.selected = subcommand
				subcommand.parent = cmd
				return subcommand.Parse(commandArgs[1:], options...)
			}
		}
	}

	// We didn't find a matching subcommand, so we selected ourselves.
	cmd.selected = cmd

	// The terminal command args may include flags, which we parse out.
	if _, err := cmd.Flags.reparse(commandArgs); err != nil {
		return fmt.Errorf("%s: reparse args: %w", cmd.Name, err)
	}

	// Parse complete.
	return nil
}

// Run the Exec function of the terminal command selected during the parse
// phase, passing the args left over after parsing. Calling [Command.Run]
// without first calling [Command.Parse] will result in [ErrNotParsed].
func (cmd *Command) Run(ctx context.Context) error {
	switch {
	case !cmd.isParsed:
		return ErrNotParsed
	case cmd.isParsed && cmd.selected == nil:
		return ErrNotParsed
	case cmd.isParsed && cmd.selected == cmd && cmd.Exec == nil:
		return fmt.Errorf("%s: %w", cmd.Name, ErrNoExec)
	case cmd.isParsed && cmd.selected == cmd && cmd.Exec != nil:
		return cmd.Exec(ctx, cmd.Flags.GetArgs())
	default:
		return cmd.selected.Run(ctx)
	}
}

// ParseAndRun calls [Command.Parse] and, upon success, [Command.Run].
func (cmd *Command) ParseAndRun(ctx context.Context, args []string, options ...Option) error {
	if err := cmd.Parse(args, options...); err != nil {
		return err
	}

	if err := cmd.Run(ctx); err != nil {
		return err
	}

	return nil
}

// GetSelected returns the terminal command selected during the parse phase, or
// nil if the command hasn't been successfully parsed.
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

// Reset every command in the command tree to its initial state, including all
// flag sets.
func (cmd *Command) Reset() error {
	if cmd.Flags != nil {
		if err := cmd.Flags.Reset(); err != nil {
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

	return nil
}
