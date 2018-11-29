package ff

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/pkg/errors"
)

// Parse the flags in the flag set from the provided (presumably commandline)
// args. Additional options may be provided to parse from a config file and/or
// environment variables in that priority order.
func Parse(fs *flag.FlagSet, args []string, options ...Option) error {
	var c Context
	for _, option := range options {
		option(&c)
	}

	if err := fs.Parse(args); err != nil {
		return errors.Wrap(err, "error parsing commandline args")
	}

	provided := map[string]bool{}
	fs.Visit(func(f *flag.Flag) {
		fmt.Fprintf(os.Stderr, "### Parse provided[%s] = true\n", f.Name)
		provided[f.Name] = true
	})

	if c.configFile == "" && c.configFileFlagName != "" {
		fs.VisitAll(func(f *flag.Flag) {
			if f.Name == c.configFileFlagName {
				c.configFile = f.Value.String()
			}
		})
	}

	if c.configFile != "" && c.configParser != nil {
		f, err := os.Open(c.configFile)
		if err != nil {
			return err
		}
		defer f.Close()

		c.configParser(f, func(name, value string) error {
			if fs.Lookup(name) == nil {
				return errors.Errorf("config file flag %q not defined in flag set", name)
			}

			if provided[name] {
				return nil // commandline args take precedence
			}

			if err := fs.Set(name, value); err != nil {
				return errors.Wrapf(err, "error setting flag %q from config file", name)
			}

			return nil
		})
	}

	fs.Visit(func(f *flag.Flag) {
		provided[f.Name] = true
	})

	if c.envVarPrefix != "" {
		var errs []string
		fs.VisitAll(func(f *flag.Flag) {
			if provided[f.Name] {
				return // commandline args and config file take precedence
			}

			var key string
			{
				key = strings.ToUpper(f.Name)
				key = envVarReplacer.Replace(key)
				key = strings.ToUpper(c.envVarPrefix) + "_" + key
			}
			if value := os.Getenv(key); value != "" {
				for _, individual := range strings.Split(value, ",") {
					if err := fs.Set(f.Name, strings.TrimSpace(individual)); err != nil {
						errs = append(errs, errors.Wrapf(err, "error setting flag %q from env var %q", f.Name, key).Error())
					}
				}
			}
		})
		if len(errs) > 0 {
			return errors.Errorf("error parsing env vars: %s", strings.Join(errs, "; "))
		}
	}

	return nil
}

// Context contains private fields used during parsing.
type Context struct {
	configFile         string
	configFileFlagName string
	configParser       Parser
	envVarPrefix       string
}

// Option controls some aspect of parse behavior.
type Option func(*Context)

// WithConfigFile tells parse to read the provided filename as a config file.
// Requires WithConfigParser, and overrides WithConfigFileFlagName.
func WithConfigFile(filename string) Option {
	return func(c *Context) {
		c.configFile = filename
	}
}

// WithConfigFileFlagName tells parse to treat the flag with the given name as a
// config file. Requires WithConfigParser, and is overridden by WithConfigFile.
func WithConfigFileFlagName(name string) Option {
	return func(c *Context) {
		c.configFileFlagName = name
	}
}

// WithConfigParser tells parse how to interpret the config file provided via
// WithConfigFileFlagName.
func WithConfigParser(p Parser) Option {
	return func(c *Context) {
		c.configParser = p
	}
}

// WithEnvVarPrefix tells parse to look in the environment for variables with
// the given prefix. Flag names are converted to environment variables by
// capitalizing and replacing separator characters like `.` or `-` with `_`.
// Additionally, if the env var value contains commas, each comma-delimited
// token is treated as a separate instance of the associated flag name.
func WithEnvVarPrefix(prefix string) Option {
	return func(c *Context) {
		c.envVarPrefix = prefix
	}
}

// Parser interprets the config file represented by the reader
// and calls the set function for each discovered flag pair.
type Parser func(r io.Reader, set func(name, value string) error) error

// PlainParser is a parser for config files in an extremely simple format. Each
// line is tokenized as a single key/value pair. The first whitespace-delimited
// token in the line is interpreted as the flag name, and all remaining tokens
// are interpreted as the value. Any leading hyphens on the flag name are
// ignored.
func PlainParser(r io.Reader, set func(name, value string) error) error {
	s := bufio.NewScanner(r)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" {
			continue // skip empties
		}

		if line[0] == '#' {
			continue // skip comments
		}

		var (
			name  string
			value string
			index = strings.IndexRune(line, ' ')
		)
		if index < 0 {
			name, value = line, "true" // boolean option
		} else {
			name, value = line[:index], strings.TrimSpace(line[index:])
		}

		if i := strings.IndexRune(value, '#'); i >= 0 {
			value = strings.TrimSpace(value[:i])
		}

		if err := set(name, value); err != nil {
			return err
		}
	}
	return nil
}

var envVarReplacer = strings.NewReplacer(
	"-", "_",
	".", "_",
	"/", "_",
)
