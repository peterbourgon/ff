# ffcli [GoDoc](https://godoc.org/github.com/peterbourgon/ff/ffcli?status.svg)](https://godoc.org/github.com/peterbourgon/ff/ffcli)

ffcli stands for flags-first command line interface, and provides an opinionated
way to build CLI tools with commands and subcommands. It's a little bit
lower-level than popular, all-in-one packages like [spf13/cobra][cobra]
[urfave/cli][urfave], or [alecthomas/kingpin][kingpin].

[cobra]: https://github.com/spf13/cobra
[urfave]: https://github.com/urfave/cli
[kingpin]: https://github.com/alecthomas/kingpin

## Usage

The core of the package is the [ffcli.Command][command]. Here is the simplest
possible example of an ffcli program.

[command]: https://godoc.org/github.com/peterbourgon/ff/ffcli#Command

```go
root := &ffcli.Command{
	Exec: func(ctx context.Context, args []string) error {
		fmt.Println("hello world")
		return nil
	},
}

root.ParseAndRun(context.Background(), os.Args)
```

Most CLIs use flags and arguments to control behavior. Here is a command which
takes a string to repeat as an argument, and the number of times to repeat it as
a flag.

```go
fs := flag.NewFlagSet("repeat", flag.ExitOnError)
n := fs.Int("n", 3, "number of repetitions")

root := &ffcli.Command{
	Usage:     "repeat [-n times] <arg>",
	ShortHelp: "Repeatedly print the argument to stdout.",
	FlagSet:   fs,
	Exec: func(ctx context.Context, args []string) error {
		if nargs := len(args); nargs != 1 {
			return fmt.Errorf("repeat requires exactly 1 argument, but you provided %d", nargs)
		}
		for i := 0; i < *n; i++ {
			fmt.Fprintf(os.Stdout, "%s\n", args[0])
		}
		return nil
	},
}

if err := root.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
	log.Fatal(err)
}
```

Each command may have subcommands, allowing you to build a command tree.

```go
var (
	rootFlagSet   = flag.NewFlagSet("textctl", flag.ExitOnError)
	verbose       = rootFlagSet.Bool("v", false, "increase log verbosity")
	repeatFlagSet = flag.NewFlagSet("textctl repeat", flag.ExitOnError)
	n             = repeatFlagSet.Int("n", 3, "how many times to repeat")
)

repeat := &ffcli.Command{
	Name:      "repeat",
	Usage:     "textctl repeat [-n times] <arg>",
	ShortHelp: "Repeatedly print the argument to stdout.",
	FlagSet:   repeatFlagSet,
	Exec:      func(_ context.Context, args []string) error { ... },
}

count := &ffcli.Command{
	Name:      "count",
	Usage:     "textctl count [<arg> ...]",
	ShortHelp: "Count the number of bytes in the arguments.",
	Exec:      func(_ context.Context, args []string) error { ... },
}

root := &ffcli.Command{
	Usage:       "textctl [flags] <subcommand>",
	FlagSet:     rootFlagSet,
	Subcommands: []*ffcli.Command{repeat, count},
	Exec:        func(context.Context, []string) error { ... },
}

if err := root.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
	log.Fatal(err)
}
```

ParseAndRun can also be split into distinct Parse and Run phases, allowing you
to perform two-phase setup or initialization of e.g. API clients based on user
configuration. 

I believe this minimal set of features, combined with the tools for abstraction
already provided by the language, are enough to express almost any commandline
application, while keeping the code significantly more understandable, testable,
and maintainable than other CLI packages and frameworks.

## Examples

See [the examples directory][examples]. If you'd like an example of a specific
type of program structure, or a CLI that satisfies a specific requirement,
please [file an issue][issue].

[examples]: https://github.com/peterbourgon/ff/tree/master/ffcli/examples
[issue]: https://github.com/peterbourgon/ff/issues/new
