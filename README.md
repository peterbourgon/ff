# ff [![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/peterbourgon/ff/v4) [![Latest Release](https://img.shields.io/github/v/release/peterbourgon/ff?style=flat-square)](https://github.com/peterbourgon/ff/releases/latest) ![Build Status](https://github.com/peterbourgon/ff/actions/workflows/test.yaml/badge.svg?branch=main)

ff is a flags-first approach to configuration.

The basic rationale is that that `myprogram -h` should reliably describe the
complete configuration "surface area" of a program. Therefore, every config
parameter should be defined as a flag. This module provides a simple and robust
way to define those flags, and to parse them from command-line arguments,
environment variables, and/or a config file.

Building a command-line application in the style of `kubectl` or `docker`?
[Command](#command) provides a declarative approach that's simpler to write, and
easier to maintain, than many common alternatives.

## Usage

This module provides a getopts(3)-inspired flag set, used as follows.

```go
fs := ff.NewFlags("myprogram")
var (
	listenAddr = fs.StringLong("listen", "localhost:8080", "listen address")
	refresh    = fs.Duration('r', "refresh", 15*time.Second, "refresh interval")
	debug      = fs.Bool('d', "debug", false, "log debug information")
	_          = fs.StringLong("config", "", "config file (optional)")
)
```

It's also possible to adapt a standard library flag set. In this case, be sure
to use the ContinueOnError error handling strategy. Other options either panic
or terminate the program on parse errors. Rude!

```go
fs := flag.NewFlagSet("myprogram", flag.ContinueOnError)
var (
	listenAddr = fs.String("listen", "localhost:8080", "listen address")
	refresh    = fs.Duration("refresh", 15*time.Second, "refresh interval")
	debug      = fs.Bool("debug", "log debug information")
	_          = fs.String("config", "", "config file (optional)")
)
```

Once you have a set of flags, you can parse them as follows; options can be
provided to control parse behavior.

```go
err := ff.Parse(fs, os.Args[1:],
	ff.WithEnvVarPrefix("MY_PROGRAM"),
	ff.WithConfigFileFlag("config"),
	ff.WithConfigFileParser(ff.PlainParser),
)
```

In the above example, flags are set from the provided command-line arguments
first; from env vars beginning with `MY_PROGRAM` next; and finally, if the user
specifies a config file, from values in that file, as parsed by PlainParser.

## Environment variables

It's possible to take runtime configuration from the environment. The options
[WithEnvVars][withenvvars] and [WithEnvVarPrefix][withenvvarprefix] enable this
feature, and determine how flag names are mapped to environment variable names.

[withenvvars]: https://pkg.go.dev/github.com/peterbourgon/ff/v4#WithEnvVars
[withenvvarprefix]: https://pkg.go.dev/github.com/peterbourgon/ff/v4#WithEnvVarPrefix

```go
fs := ff.NewFlags("myservice")
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

It's possible to take runtime configuration from config files. The options
[WithConfigFile][withconfigfile], [WithConfigFileFlag][withconfigfileflag], and
[WithConfigFileParser][withconfigfileparser] control how config files are
specified and parsed. This module includes support for JSON, YAML, TOML, and
.env config files, as well as the simple [PlainParser][plainparser] format.

[withconfigfile]: https://pkg.go.dev/github.com/peterbourgon/ff/v4#WithConfigFile
[withconfigfileflag]: https://pkg.go.dev/github.com/peterbourgon/ff/v4#WithConfigFileFlag
[withconfigfileparser]: https://pkg.go.dev/github.com/peterbourgon/ff/v4#WithConfigFileParser
[plainparser]: https://pkg.go.dev/github.com/peterbourgon/ff/v4#PlainParser

```go
fs := ff.NewFlags("myservice")
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

Command-line args have the highest priority, because they're explicitly given to
each running instance of a program by the user. Think of command-line args as the
"user" configuration.

Environment variables have the next-highest priority, because they reflect
configuration set in the runtime context. Think of env vars as the "session"
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
