# ff [![Latest Release](https://img.shields.io/github/release/peterbourgon/ff.svg?style=flat-square)](https://github.com/peterbourgon/ff/releases/latest) [![GoDoc](https://godoc.org/github.com/peterbourgon/ff?status.svg)](https://godoc.org/github.com/peterbourgon/ff) [![Travis CI](https://travis-ci.org/peterbourgon/ff.svg?branch=master)](https://travis-ci.org/peterbourgon/ff)

ff stands for flags-first, and provides an opinionated way to populate a
[flag.FlagSet](https://golang.org/pkg/flag#FlagSet) with configuration data from
the environment. Specifically, it allows data to be parsed from commandline
args, a configuration file, and environment variables, in that priority order.

## Usage

Define a flag.FlagSet in your func main.

```go
func main() {
	fs := flag.NewFlagSet("my-program", flag.ExitOnError)
	var (
		listenAddr = fs.String("listen-addr", "localhost:8080", "listen address")
		refresh    = fs.Duration("refresh", 15*time.Second, "refresh interval")
		debug      = fs.Bool("debug", false, "log debug information")
		_          = fs.String("config", "", "config file (optional)")
	)
```

Then, call ff.Parse instead of fs.Parse.

```go
	ff.Parse(fs, os.Args[1:],
		ff.WithConfigFileFlag("config"),
		ff.WithConfigFileParser(ff.PlainParser),
		ff.WithEnvVarPrefix("MY_PROGRAM"),
	)
```

This example will parse flags from the commandline args, just like regular
package flag, with the highest priority. If a `-config` file is specified, it
will try to parse it using the PlainParser, which expects files in this format:

```
listen-addr localhost:8080
refresh 30s
debug true
```

It's simple to write your own config file parser.

```go
// ConfigFileParser interprets the config file represented by the reader
// and calls the set function for each parsed flag pair.
type ConfigFileParser func(r io.Reader, set func(name, value string) error) error
```

Finally, it will look in the environment for variables with a `MY_PROGRAM`
prefix. Flag names are capitalized, and separator characters are converted to
underscores. In this case, for example, `MY_PROGRAM_LISTEN_ADDR` would match to
`listen-addr`.
