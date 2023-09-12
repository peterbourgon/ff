package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffhelp"
)

func main() {
	fs := ff.NewFlagSet("basicflags")
	var (
		config  = fs.String('c', "config", "", "config file")
		delta   = fs.Duration('d', "delta", time.Second, "value for `∆` parameter")
		epsilon = fs.IntLong("epsilon", 32, "value for `ε` parameter")
		urls    = fs.StringSet('u', "url", "remote URL (repeatable)")
		verbose = fs.Bool('v', "verbose", "verbose logging")
	)

	err := ff.Parse(fs, os.Args[1:],
		ff.WithEnvVarPrefix("BASICFLAGS"), // try `env BASICFLAGS_DELTA=33ms basicflags`
		ff.WithConfigFileFlag("config"),   // try providing a file with `delta 33ms`
		ff.WithConfigFileParser(ff.PlainParser),
	)
	switch {
	case errors.Is(err, ff.ErrHelp):
		fmt.Fprintf(os.Stderr, "%s\n", ffhelp.Flags(fs))
		os.Exit(0)
	case err != nil:
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("config: %q\n", *config)
	fmt.Printf("delta: %s\n", *delta)
	fmt.Printf("epsilon: %d\n", *epsilon)
	fmt.Printf("urls: %v\n", *urls)
	fmt.Printf("verbose: %v\n", *verbose)
	fmt.Printf("fs.GetArgs: %v\n", fs.GetArgs())
}
