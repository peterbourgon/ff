package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffhelp"
)

// textctl is a simple application where all commands are built up in func main.
// It demonstrates how to declare commands, how to wire them into a command
// tree, and how to give subcommands access to parent command flags.

func main() {
	rootFlags := ff.NewFlags("textctl")
	verbose := rootFlags.Bool('v', "verbose", "increase log verbosity")
	rootCmd := &ff.Command{
		Name:  "textctl",
		Usage: "textctl [FLAGS] <SUBCOMMAND>",
		Flags: rootFlags,
	}

	repeatFlags := ff.NewFlags("repeat").SetParent(rootFlags) // SetParent allows repeatFlags access to rootFlags
	n := repeatFlags.IntShort('n', 3, "how many times to repeat")
	repeatCmd := &ff.Command{
		Name:      "repeat",
		Usage:     "textctl repeat [-n TIMES] <ARG>",
		ShortHelp: "repeatedly print the argument to stdout",
		Flags:     repeatFlags,
		Exec: func(_ context.Context, args []string) error { // defining Exec inline allows it to access the e.g. verbose flag, above
			if n := len(args); n != 1 {
				return fmt.Errorf("repeat requires exactly 1 argument, but you provided %d", n)
			}
			if *verbose {
				fmt.Fprintf(os.Stderr, "repeat: n=%d\n", *n)
				fmt.Fprintf(os.Stderr, "repeat: will generate %dB of output\n", (*n)*len(args[0]))
			}
			for i := 0; i < *n; i++ {
				fmt.Fprintf(os.Stdout, "%s\n", args[0])
			}
			return nil
		},
	}
	rootCmd.Subcommands = append(rootCmd.Subcommands, repeatCmd) // add the repeat command underneath the root comand

	countCmd := &ff.Command{
		Name:      "count",
		Usage:     "textctl count [<ARG> ...]",
		ShortHelp: "count the number of bytes in the arguments",
		Flags:     ff.NewFlags("count").SetParent(rootFlags), // count has no flags itself, but it should still be able to parse root flags
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
	rootCmd.Subcommands = append(rootCmd.Subcommands, countCmd) // add the count command underneath the root command

	err := rootCmd.ParseAndRun(context.Background(), os.Args[1:])
	if errors.Is(err, ff.ErrHelp) || errors.Is(err, ff.ErrNoExec) {
		fmt.Fprintf(os.Stderr, "\n%s\n", ffhelp.Command(rootCmd))
		os.Exit(0)
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}
