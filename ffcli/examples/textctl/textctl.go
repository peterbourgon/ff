package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/peterbourgon/ff/v3/ffcli"
)

// textctl is a simple applications in which all commands are built up in func
// main. It demonstrates how to declare minimal commands, how to wire them
// together into a command tree, and one way to allow subcommands access to
// flags set in parent commands.

func main() {
	var (
		rootFlagSet   = flag.NewFlagSet("textctl", flag.ExitOnError)
		verbose       = rootFlagSet.Bool("v", false, "increase log verbosity")
		repeatFlagSet = flag.NewFlagSet("textctl repeat", flag.ExitOnError)
		n             = repeatFlagSet.Int("n", 3, "how many times to repeat")
	)

	repeat := &ffcli.Command{
		Name:       "repeat",
		ShortUsage: "textctl repeat [-n times] <arg>",
		ShortHelp:  "Repeatedly print the argument to stdout.",
		FlagSet:    repeatFlagSet,
		Exec: func(_ context.Context, args []string) error {
			if n := len(args); n != 1 {
				return fmt.Errorf("repeat requires exactly 1 argument, but you provided %d", n)
			}
			if *verbose {
				fmt.Fprintf(os.Stderr, "repeat: will generate %dB of output\n", (*n)*len(args[0]))
			}
			for i := 0; i < *n; i++ {
				fmt.Fprintf(os.Stdout, "%s\n", args[0])
			}
			return nil
		},
	}

	count := &ffcli.Command{
		Name:       "count",
		ShortUsage: "count [<arg> ...]",
		ShortHelp:  "Count the number of bytes in the arguments.",
		Exec: func(_ context.Context, args []string) error {
			if *verbose {
				fmt.Fprintf(os.Stderr, "count: argument count %d\n", len(args))
			}
			var n int
			for _, arg := range args {
				n += len(arg)
			}
			fmt.Fprintf(os.Stdout, "%d\n", n)
			return nil
		},
	}

	root := &ffcli.Command{
		ShortUsage:  "textctl [flags] <subcommand>",
		FlagSet:     rootFlagSet,
		Subcommands: []*ffcli.Command{repeat, count},
		Exec: func(context.Context, []string) error {
			return flag.ErrHelp
		},
	}

	if err := root.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
