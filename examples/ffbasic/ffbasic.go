package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffhelp"
)

func main() {
	fs := ff.NewFlags("ffbasic").SetDescription("ffbasic -- an example program")
	var (
		foo   = fs.String('f', "foo", "", "some string value")
		bar   = fs.IntLong("bar", 123, "some int value")
		debug = fs.Bool('d', "debug", false, "debug logging")
	)

	err := ff.Parse(fs, os.Args[1:])
	switch {
	case errors.Is(err, ff.ErrHelp):
		fmt.Fprintf(os.Stderr, "%s\n", ffhelp.Flags(fs))
		os.Exit(0)
	case err != nil:
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("foo:   %q\n", *foo)
	fmt.Printf("bar:   %d\n", *bar)
	fmt.Printf("debug: %v\n", *debug)
	fmt.Printf("args:  %v\n", fs.GetArgs())
}
