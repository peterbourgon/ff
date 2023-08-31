package ff_test

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/peterbourgon/ff/v4"
)

func ExampleParse_args() {
	fs := ff.NewFlags("myprogram")
	var (
		listen  = fs.StringLong("listen", "localhost:8080", "listen address")
		refresh = fs.Duration('r', "refresh", 15*time.Second, "refresh interval")
		debug   = fs.Bool('d', "debug", false, "log debug information")
	)

	err := ff.Parse(fs, []string{"--refresh=1s", "-d"})

	fmt.Printf("err=%v\n", err)
	fmt.Printf("listen=%v\n", *listen)
	fmt.Printf("refresh=%v\n", *refresh)
	fmt.Printf("debug=%v\n", *debug)

	// Output:
	// err=<nil>
	// listen=localhost:8080
	// refresh=1s
	// debug=true
}

func ExampleParse_env() {
	fs := ff.NewFlags("myprogram")
	var (
		listen  = fs.StringLong("listen", "localhost:8080", "listen address")
		refresh = fs.Duration('r', "refresh", 15*time.Second, "refresh interval")
		debug   = fs.Bool('d', "debug", false, "log debug information")
	)

	os.Setenv("MY_PROGRAM_REFRESH", "3s")

	err := ff.Parse(fs, []string{},
		ff.WithEnvVarPrefix("MY_PROGRAM"),
	)

	fmt.Printf("err=%v\n", err)
	fmt.Printf("listen=%v\n", *listen)
	fmt.Printf("refresh=%v\n", *refresh)
	fmt.Printf("debug=%v\n", *debug)

	// Output:
	// err=<nil>
	// listen=localhost:8080
	// refresh=3s
	// debug=false
}

func ExampleParse_config() {
	fs := ff.NewFlags("myprogram")
	var (
		listen  = fs.StringLong("listen", "localhost:8080", "listen address")
		refresh = fs.Duration('r', "refresh", 15*time.Second, "refresh interval")
		debug   = fs.Bool('d', "debug", false, "log debug information")
		_       = fs.String('c', "config", "", "path to config file")
	)

	f, _ := os.CreateTemp("", "ExampleParse_config")
	defer func() { f.Close(); os.Remove(f.Name()) }()
	fmt.Fprint(f, `
		debug
		listen localhost:9999
	`)

	err := ff.Parse(fs, []string{"-c", f.Name()},
		ff.WithConfigFileFlag("config"),
		ff.WithConfigFileParser(ff.PlainParser),
	)

	fmt.Printf("err=%v\n", err)
	fmt.Printf("listen=%v\n", *listen)
	fmt.Printf("refresh=%v\n", *refresh)
	fmt.Printf("debug=%v\n", *debug)

	// Output:
	// err=<nil>
	// listen=localhost:9999
	// refresh=15s
	// debug=true
}

func ExampleParse_stdlib() {
	fs := flag.NewFlagSet("myprogram", flag.ContinueOnError)
	var (
		listen  = fs.String("listen", "localhost:8080", "listen address")
		refresh = fs.Duration("refresh", 15*time.Second, "refresh interval")
		debug   = fs.Bool("debug", false, "log debug information")
	)

	err := ff.Parse(fs, []string{"--debug", "-refresh=2s", "-listen", "localhost:9999"})

	fmt.Printf("err=%v\n", err)
	fmt.Printf("listen=%v\n", *listen)
	fmt.Printf("refresh=%v\n", *refresh)
	fmt.Printf("debug=%v\n", *debug)

	// Output:
	// err=<nil>
	// listen=localhost:9999
	// refresh=2s
	// debug=true
}

func ExampleParse_help() {
	fs := ff.NewFlags("myprogram")
	var (
		listen  = fs.StringLong("listen", "localhost:8080", "listen address")
		refresh = fs.DurationLong("refresh", 15*time.Second, "refresh interval")
		debug   = fs.BoolLong("debug", false, "log debug information")
	)

	err := ff.Parse(fs, []string{"-h"})

	fmt.Printf("err=%v\n", err)
	fmt.Printf("listen=%v\n", *listen)
	fmt.Printf("refresh=%v\n", *refresh)
	fmt.Printf("debug=%v\n", *debug)

	// Output:
	// err=parse args: flag: help requested
	// listen=localhost:8080
	// refresh=15s
	// debug=false
}
