# ff 

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/peterbourgon/ff/v4) 
[![Latest Release](https://img.shields.io/github/v/release/peterbourgon/ff?style=flat-square)](https://github.com/peterbourgon/ff/releases/latest) 
![Build Status](https://github.com/peterbourgon/ff/actions/workflows/test.yaml/badge.svg?branch=main)

ff is a flags-first approach to configuration.

The basic idea is that `myprogram -h` should always show the complete
configuration "surface area" of a program. Therefore, every config parameter
should be defined as a flag. This module provides a simple and robust way to
define those flags, and to parse them from command-line arguments, environment
variables, and/or config files.

Building a command-line application in the style of `kubectl` or `docker`?
[ff.Command](#ffcommand) offers a declarative approach that may be simpler to
write, and easier to maintain, than many common alternatives.

## Note

This README describes the pre-release version v4 of ff. For the stable version,
see [ff/v3](https://pkg.go.dev/github.com/peterbourgon/ff/v3).

## Usage

### Parse a flag.FlagSet

Parse a flag.FlagSet from commandline args, env vars, and/or a config file, by
using [ff.Parse][ffparse] instead of flag.FlagSet.Parse. Use
[options][option] to control parse behavior.

[ffparse]: https://pkg.go.dev/github.com/peterbourgon/ff/v4#Parse
[option]: https://pkg.go.dev/github.com/peterbourgon/ff/v4#Option

```go
fs := flag.NewFlagSet("myprogram", flag.ContinueOnError)
var (
	listenAddr = fs.String("listen", "localhost:8080", "listen address")
	refresh    = fs.Duration("refresh", 15*time.Second, "refresh interval")
	debug      = fs.Bool("debug", false, "log debug information")
	_          = fs.String("config", "", "config file (optional)")
)

ff.Parse(fs, os.Args[1:],
	ff.WithEnvVarPrefix("MY_PROGRAM"),
	ff.WithConfigFileFlag("config"),
	ff.WithConfigFileParser(ff.PlainParser),
)

fmt.Printf("listen=%s refresh=%s debug=%v\n", *listen, *refresh, *debug)
```

```shell
$ myprogram -listen=localhost:9090
listen=localhost:9090 refresh=15s debug=false

$ env MY_PROGRAM_DEBUG=1 myprogram
listen=localhost:8080 refresh=15s debug=true

$ printf 'refresh 30s \n debug \n' > my.conf
$ myprogram -config=my.conf
listen=localhost:8080 refresh=30s debug=true
```

### Upgrade to an ff.FlagSet

Alternatively, you can use the getopts(3)-inspired [ff.FlagSet][flagset], which
provides short (-f) and long (--foo) flag names, more useful flag types, and
other niceities.

[flagset]: https://github.com/peterbourgon/ff/v4#FlagSet

```go
fs := ff.NewFlagSet("myprogram")
var (
	addrs     = fs.StringSet('a', "addr", "remote address (repeatable)")
	compress  = fs.Bool('c', "compress", "enable compression")
	transform = fs.Bool('t', "transform", "enable transformation")
	loglevel  = fs.StringEnum('l', "log", "log level: debug, info, error", "info", "debug", "error")
	_         = fs.StringLong("config", "", "config file (optional)")
)

ff.Parse(fs, os.Args[1:],
	ff.WithEnvVarPrefix("MY_PROGRAM"),
	ff.WithConfigFileFlag("config"),
	ff.WithConfigFileParser(ff.PlainParser),
)

fmt.Printf("addrs=%v compress=%v transform=%v loglevel=%v\n", *addrs, *compress, *transform, *loglevel)
```

```shell
$ env MY_PROGRAM_LOG=debug myprogram -afoo -a bar --addr=baz --addr qux -ct
addrs=[foo bar baz qux] compress=true transform=true loglevel=debug
```

### Parent flag sets

ff.FlagSet supports the notion of a parent flag set, which allows a "child" flag
set to parse all "parent" flags, in addition to their own flags.

```go
parentfs := ff.NewFlagSet("parentcommand")
var (
	loglevel = parentfs.StringEnum('l', "log", "log level: debug, info, error", "info", "debug", "error")
	_        = parentfs.StringLong("config", "", "config file (optional)")
)

childfs := ff.NewFlagSet("childcommand").SetParent(parentfs)
var (
	compress  = childfs.Bool('c', "compress", "enable compression")
	transform = childfs.Bool('t', "transform", "enable transformation")
	refresh   = childfs.DurationLong("refresh", 15*time.Second, "refresh interval")
)

ff.Parse(childfs, os.Args[1:],
	ff.WithEnvVarPrefix("MY_PROGRAM"),
	ff.WithConfigFileFlag("config"),
	ff.WithConfigFileParser(ff.PlainParser),
)

fmt.Printf("loglevel=%v compress=%v transform=%v refresh=%v\n", *loglevel, *compress, *transform, *refresh)
```

```shell
$ myprogram --log=debug --refresh=1s
loglevel=debug compress=false transform=false refresh=1s

$ printf 'log error \n refresh 5s \n' > my.conf
$ myprogram --config my.conf
loglevel=error compress=false transform=false refresh=5s
```

### Help output

Unlike flag.FlagSet, the ff.FlagSet doesn't emit help text to os.Stderr as an
invisible side effect of a failed parse. When using an ff.FlagSet, callers are
expected to check the error returned by parse, and to emit help text to the user
as appropriate. [Package ffhelp][ffhelp] provides functions that produce help
text in a standard format, and tools for creating your own help text format.

[ffhelp]: https://pkg.go.dev/github.com/peterbourgon/ff/v4/ffhelp

```go
parentfs := ff.NewFlagSet("parentcommand")
var (
	loglevel  = parentfs.StringEnum('l', "log", "log level: debug, info, error", "info", "debug", "error")
	_         = parentfs.StringLong("config", "", "config file (optional)")
)

childfs := ff.NewFlagSet("childcommand").SetParent(parentfs)
var (
	compress  = childfs.Bool('c', "compress", "enable compression")
	transform = childfs.Bool('t', "transform", "enable transformation")
	refresh   = childfs.DurationLong("refresh", 15*time.Second, "refresh interval")
)

if err := ff.Parse(childfs, os.Args[1:],
	ff.WithEnvVarPrefix("MY_PROGRAM"),
	ff.WithConfigFileFlag("config"),
	ff.WithConfigFileParser(ff.PlainParser),
); err != nil {
	fmt.Printf("%s\n", ffhelp.Flags(childfs))
	fmt.Printf("err=%v\n", err)
} else {
	fmt.Printf("loglevel=%v compress=%v transform=%v refresh=%v\n", *loglevel, *compress, *transform, *refresh)
}
```

```shell
$ childcommand -h
NAME
  childcommand

FLAGS (childcommand)
  -c, --compress           enable compression
  -t, --transform          enable transformation
      --refresh DURATION   refresh interval (default: 15s)

FLAGS (parentcommand)
  -l, --log STRING         log level: debug, info, error (default: info)
      --config STRING      config file (optional)

err=parse args: flag: help requested
```

## Parse priority

Command-line args have the highest priority, because they're explicitly provided
to the program by the user. Think of command-line args as the "user"
configuration.

Environment variables have the next-highest priority, because they represent
configuration in the runtime environment. Think of env vars as the "session"
configuration.

Config files have the lowest priority, because they represent config that's
static to the host. Think of config files as the "host" configuration.

## ff.Command

[ff.Command][command] is a tool for building larger CLI programs with
sub-commands, like `docker` or `kubectl`. It's a declarative and lightweight
alternative to more common frameworks like [spf13/cobra][cobra],
[urfave/cli][urfave], or [alecthomas/kingpin][kingpin].

[command]: https://pkg.go.dev/github.com/peterbourgon/ff/v4#Command
[cobra]: https://github.com/spf13/cobra
[urfave]: https://github.com/urfave/cli
[kingpin]: https://github.com/alecthomas/kingpin

Commands are concerned only with the core mechanics of defining a command tree,
parsing flags, and selecting a command to run. They're not intended to be a
one-stop-shop for everything a command-line application may need. Features like
tab completion, colorized output, etc. are orthogonal to command tree parsing,
and can be easily added on top.

Here's a simple example of a basic command tree.

```go
// textctl -- root command
textctlFlags := ff.NewFlagSet("textctl")
verbose := textctlFlags.Bool('v', "verbose", "increase log verbosity")
textctlCmd := &ff.Command{
	Name:  "textctl",
	Usage: "textctl [FLAGS] SUBCOMMAND ...",
	Flags: textctlFlags,
}

// textctl repeat -- subcommand
repeatFlags := ff.NewFlagSet("repeat").SetParent(textctlFlags) // <-- set parent flag set
n := repeatFlags.IntShort('n', 3, "how many times to repeat")
repeatCmd := &ff.Command{
	Name:      "repeat",
	Usage:     "textctl repeat [-n TIMES] ARG",
	ShortHelp: "repeatedly print the first argument to stdout",
	Flags:     repeatFlags,
	Exec:      func(ctx context.Context, args []string) error { /* ... */ },
}
textctlCmd.Subcommands = append(textctlCmd.Subcommands, repeatCmd) // <-- append to parent subcommands

// ...

if err := textctlCmd.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
	fmt.Fprintf(os.Stderr, "%s\n", ffhelp.Command(textctlCmd))
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}
```

More sophisticated programs are available in [the examples directory][examples].

[examples]: https://github.com/peterbourgon/ff/tree/main/examples
