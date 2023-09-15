package ff_test

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffhelp"
	"github.com/peterbourgon/ff/v4/ffval"
)

func ExampleParse_args() {
	fs := ff.NewFlagSet("myprogram")
	var (
		listen  = fs.StringLong("listen", "localhost:8080", "listen address")
		refresh = fs.Duration('r', "refresh", 15*time.Second, "refresh interval")
		debug   = fs.Bool('d', "debug", "log debug information")
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
	fs := ff.NewFlagSet("myprogram")
	var (
		listen  = fs.StringLong("listen", "localhost:8080", "listen address")
		refresh = fs.Duration('r', "refresh", 15*time.Second, "refresh interval")
		debug   = fs.Bool('d', "debug", "log debug information")
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
	fs := ff.NewFlagSet("myprogram")
	var (
		listen  = fs.StringLong("listen", "localhost:8080", "listen address")
		refresh = fs.Duration('r', "refresh", 15*time.Second, "refresh interval")
		debug   = fs.Bool('d', "debug", "log debug information")
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
	fs := ff.NewFlagSet("myprogram")
	var (
		listen  = fs.StringLong("listen", "localhost:8080", "listen address")
		refresh = fs.DurationLong("refresh", 15*time.Second, "refresh interval")
		debug   = fs.BoolLong("debug", "log debug information")
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

func ExampleFlagSet_AddStruct() {
	var firstFlags struct {
		Alpha   string `ff:"shortname: a, longname: alpha, usage: alpha string,    default: abc   "`
		Beta    int    `ff:"              longname: beta,  usage: 'beta: an int',  placeholder: β "`
		Delta   bool   `ff:"shortname: d,                  usage: 'delta, a bool', nodefault      "`
		Epsilon bool   `ff:"short: e,     long: epsilon,   usage: epsilon bool,    nodefault      "`
	}

	var secondFlags struct {
		Gamma string          `ff:" short=g | long=gamma |              | usage: gamma string       "`
		Iota  float64         `ff:"         | long=iota  | default=0.43 | usage: 🦊                 "`
		Kappa ffval.StringSet `ff:" short=k | long=kappa |              | usage: kappa (repeatable) "`
	}

	fs := ff.NewFlagSet("mycommand")
	fs.AddStruct(&firstFlags)
	fs.AddStruct(&secondFlags)
	fmt.Print(ffhelp.Flags(fs))

	// Output:
	// NAME
	//   mycommand
	//
	// FLAGS
	//   -a, --alpha STRING   alpha string (default: abc)
	//       --beta β         beta: an int (default: 0)
	//   -d                   delta, a bool
	//   -e, --epsilon        epsilon bool
	//   -g, --gamma STRING   gamma string
	//       --iota FLOAT64   🦊 (default: 0.43)
	//   -k, --kappa STRING   kappa (repeatable)
}
