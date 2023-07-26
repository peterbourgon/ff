# ff [![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/peterbourgon/ff/v4) [![Latest Release](https://img.shields.io/github/v/release/peterbourgon/ff?style=flat-square)](https://github.com/peterbourgon/ff/releases/latest) ![Build Status](https://github.com/peterbourgon/ff/actions/workflows/test.yaml/badge.svg?branch=main)

ff is a flags-first approach for programs to receive runtime configuration.

As the name suggests, it's all based on flags. Every config parameter is
expected to be defined as a flag, to ensure that `myprogram -h` will reliably
describe the complete configuration surface area of the program.

Building a command-line application in the style of `kubectl` or `docker`?
[Command](#command) offers a declarative approach that may be simpler and easier
to maintain than common alternatives.

## Usage

This module provides a getopts(3)-style flag set, which can be used as follows.

```go
fs := ff.NewFlags("myprogram")
var (
	listenAddr = fs.StringLong("listen", "localhost:8080", "listen address")
	refresh    = fs.Duration('r', "refresh", 15*time.Second, "refresh interval")
	debug      = fs.Bool('d', "debug", "log debug information")
	_          = fs.StringLong("config", "", "config file (optional)")
)
```

You can also use a standard library flag set. If you do, be sure to use the
ContinueOnError error handling strategy. Other options either panic or terminate
the program on parse errors. Rude!

```go
fs := flag.NewFlagSet("myprogram", flag.ContinueOnError)
var (
	listenAddr = fs.String("listen", "localhost:8080", "listen address")
	refresh    = fs.Duration("refresh", 15*time.Second, "refresh interval")
	debug      = fs.Bool("debug", "log debug information")
	_          = fs.String("config", "", "config file (optional)")
)
```

Once you have a set of flags, use ff.Parse to parse it. Options can be provided
to influence parsing behavior.

```go
err := ff.Parse(fs, os.Args[1:],
	ff.WithEnvVarPrefix("MY_PROGRAM"),
	ff.WithConfigFileFlag("config"),
	ff.WithConfigFileParser(ff.PlainParser),
)
```

Flags are always set from the provided command-line arguments first. In the
above example, flags will also be set from env vars beginning with `MY_PROGRAM`.
Finally, if the user specifies a config file, flags will be set from values in
that file, as parsed by ff.PlainParser.

## Environment variables

It's possible to take runtime configuration from env vars. The options
[WithEnvVars][withenvvars] and [WithEnvVarPrefix][withenvvarprefix] enable this
feature and determine how env var keys are mapped to flag names.

[withenvvars]: https://pkg.go.dev/github.com/peterbourgon/ff/v4#WithEnvVars
[withenvvarprefix]: https://pkg.go.dev/github.com/peterbourgon/ff/v4#WithEnvVarPrefix

```go
fs := flag.NewFlagSet("myservice", flag.ContinueOnError)
var (
	port  = fs.Int("port", 8080, "listen port for server (also via PORT)")
	debug = fs.Bool("debug", false, "log debug information (also via DEBUG)")
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

It's also possible to take runtime configuration from config files. The options
[WithConfigFile][withconfigfile], [WithConfigFileFlag][withconfigfileflag], and
[WithConfigFileParser][withconfigfileparser] control how config files are
specified and parsed. This module includes support for JSON, YAML, TOML, and
.env config files, and also defines its own simple config file format.

[withconfigfile]: https://pkg.go.dev/github.com/peterbourgon/ff/v4#WithConfigFile
[withconfigfileflag]: https://pkg.go.dev/github.com/peterbourgon/ff/v4#WithConfigFileFlag
[withconfigfileparser]: https://pkg.go.dev/github.com/peterbourgon/ff/v4#WithConfigFileParser

```go
fs := flag.NewFlagSet("myservice", flag.ContinueOnError)
var (
	port  = fs.Int("port", 8080, "listen port for server (also via PORT)")
	debug = fs.Bool("debug", false, "log debug information (also via DEBUG)")
	_     = fs.String("config", "", "config file")
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

Command-line args have the highest priority, because they're explicitly given to
each running instance of a program by the user -- we call command-line args the
"user" configuration.

Envioronment variables have the next-highest priority, because they reflect
configuration set in the runtime context -- we call env vars the "session"
configuration.

Config files have the lowest priority, because they represent config that's
static to the host -- we call config files the "host" configuration.

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

See [the examples directory](examples/) for sample CLI applications.
