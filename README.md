# ff [![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/peterbourgon/ff/v4) [![Latest Release](https://img.shields.io/github/v/release/peterbourgon/ff?style=flat-square)](https://github.com/peterbourgon/ff/releases/latest) ![Build Status](https://github.com/peterbourgon/ff/actions/workflows/test.yaml/badge.svg?branch=main)

ff is a flags-first approach to configuration.

The basic idea is that `myprogram -h` should always show the complete
configuration "surface area" of a program. Therefore, every config parameter
should be defined as a flag. This module provides a simple and robust way to
define those flags, and to parse them from command-line arguments, environment
variables, and/or config files.

Building a command-line application in the style of `kubectl` or `docker`?
[Command](#command) provides a declarative approach that's simpler to write, and
easier to maintain, than many common alternatives.

## Usage

Define all of the configuration parameters for your program in a flag set. This
module provides a getopts(3)-inspired [FlagSet][flagset] type, which is a
reasonable default flag set implementaiton.

[flagset]: https://pkg.go.dev/github.com/peterbourgon/ff/v4#FlagSet

```go
fs := ff.NewFlagSet("myprogram")
var (
	listenAddr = fs.StringLong("listen", "localhost:8080", "listen address")
	refresh    = fs.Duration('r', "refresh", 15*time.Second, "refresh interval")
	debug      = fs.Bool('d', "debug", false, "log debug information")
	_          = fs.StringLong("config", "", "config file (optional)")
)
```

You can also use a standard library flag.FlagSet. Be sure to pass the
ContinueOnError error handling strategy, as other options either panic or
terminate the program on parse errors -- rude!

```go
fs := flag.NewFlagSet("myprogram", flag.ContinueOnError)
var (
	listenAddr = fs.String("listen", "localhost:8080", "listen address")
	refresh    = fs.Duration("refresh", 15*time.Second, "refresh interval")
	debug      = fs.Bool("debug", "log debug information")
	_          = fs.String("config", "", "config file (optional)")
)
```

You can also use your own implementation of the [Flags][flags] interface.

[flags]: https://pkg.go.dev/github.com/peterbourgon/ff/v4#Flags

Then, use [ff.Parse][parse] to parse the flag set. [Options][options] can be
provided to control parsing behavior.

[parse]: https://pkg.go.dev/github.com/peterbourgon/ff/v4#Parse
[options]: https://pkg.go.dev/github.com/peterbourgon/ff/v4#Option

```go
err := ff.Parse(fs, os.Args[1:],
	ff.WithEnvVarPrefix("MY_PROGRAM"),
	ff.WithConfigFileFlag("config"),
	ff.WithConfigFileParser(ff.PlainParser),
)
```

Here, flags are set from the provided command-line arguments, from env vars
beginning with `MY_PROGRAM`, and, if the user specifies a config file, from
values in that file, as parsed by [ff.PlainParser][plainparser].

[plainparser]: https://pkg.go.dev/github.com/peterbourgon/ff/v4#PlainParser

Unlike other flag packages, help text is not automatically printed as a side
effect of parsing. When a user requests help via e.g. -h or --help, it's
reported as a parse error. Callers are responsible for checking parse errors,
and printing help text when appropriate. [package ffhelp][ffhelp] provides
helpers for producing help text in a standard format, but you can always write
your own.

[ffhelp]: https://pkg.go.dev/github.com/peterbourgon/ff/v4/ffhelp

```go
if errors.Is(err, ff.ErrHelp) {
	fmt.Fprintln(os.Stderr, ffhelp.Flags(fs))
	os.Exit(0)
} else if err != nil {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}
```

## Environment variables

It's possible to set flags from env vars. The options [WithEnvVars][withenvvars]
and [WithEnvVarPrefix][withenvvarprefix] enable this feature, and determine how
environment variable names map to flag names.

[withenvvars]: https://pkg.go.dev/github.com/peterbourgon/ff/v4#WithEnvVars
[withenvvarprefix]: https://pkg.go.dev/github.com/peterbourgon/ff/v4#WithEnvVarsPrefix

```go
fs := ff.NewFlagSet("myservice")
var (
	port  = fs.Int('p', "port", 8080, "listen port for server (also via PORT)")
	debug = fs.Bool('d', "debug", false, "log debug information (also via DEBUG)")
)
ff.Parse(fs, os.Args[1:], ff.WithEnvVars())
fmt.Printf("port %d, debug %v\n", *port, *debug)
```

```shell
$ env PORT=9090 myservice
port 9090, debug false
$ env PORT=9090 DEBUG=1 myservice --port=1234
port 1234, debug true
```

## Config files

It's possible to set flags from config files. The [WithConfigFileFlag][cfflag]
and [WithConfigFileParser][cfparser] options control how config files are
specified and parsed. This module includes support for JSON, YAML, TOML, and
.env config files, as well as the simple [PlainParser][plainparser] format.

[cfflag]: https://pkg.go.dev/github.com/peterbourgon/ff/v4#WithConfigFileFlag
[cfparser]: https://pkg.go.dev/github.com/peterbourgon/ff/v4#WithConfigFileParser
[plainparser]: https://pkg.go.dev/github.com/peterbourgon/ff/v4#PlainParser

```go
fs := ff.NewFlagSet("myservice")
var (
	port  = fs.IntLong("port", 8080, "listen port for server")
	debug = fs.BoolLong("debug", false, "log debug information")
	_     = fs.StringLong("config", "", "config file")
)
ff.Parse(fs, os.Args[1:], ff.WithConfigFileFlag("config"), ff.WithConfigFileParser(ff.PlainParser))
fmt.Printf("port %d, debug %v\n", *port, *debug)
```

```shell
$ printf "port 9090\n" >1.conf ; myservice --config=1.conf
port 9090, debug false
$ printf "port 9090\ndebug\n" >2.conf ; myservice --config=2.conf --port=1234
port 1234, debug true
```

## Priority

Command-line args have the highest priority, because they're explicitly provided
to the program by the user. Think of command-line args as the "user"
configuration.

Environment variables have the next-highest priority, because they represent
configuration in the runtime environment. Think of env vars as the "session"
configuration.

Config files have the lowest priority, because they represent config that's
static to the host. Think of config files as the "host" configuration.

# Commands

[Command][command] is a declarative and lightweight alternative to common CLI
frameworks like [spf13/cobra][cobra], [urfave/cli][urfave], or
[alecthomas/kingpin][kingpin].

[command]: https://pkg.go.dev/github.com/peterbourgon/ff/v4#Command
[cobra]: https://github.com/spf13/cobra
[urfave]: https://github.com/urfave/cli
[kingpin]: https://github.com/alecthomas/kingpin

Those frameworks have relatively large APIs, in order to support a large number
of "table stakes" features. In contrast, the command API is quite small, with
the immediate goal of being intuitive and productive, and the long-term goal of
producing CLI applications that are substantially easier to understand and
maintain.

Commands are concerned only with the core mechanics of defining a command tree,
parsing flags, and selecting a command to run. They're not intended to be a
one-stop-shop for everything a command-line application may need. Features like
tab completion, colorized output, etc. are orthogonal to command tree parsing,
and can be easily provided by the consumer.

See [the examples directory](examples/) for some CLI tools built with commands.
