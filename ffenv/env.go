package ffenv

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"strings"
)

// Parser is a parser for .env file format: flag=value. Each
// line is tokenized as a single key/value pair.
func Parser(fs *flag.FlagSet) func(io.Reader, func(string, string) error) error {
	return func(r io.Reader, set func(name, value string) error) error {
		return parse(fs)("", r, set)
	}
}

// Parser is a parser for config files in an extremely simple format. Each
// line is tokenized as a single key/value pair. The first =-delimited
// token in the line is interpreted as the flag name, and all remaining tokens
// are interpreted as the value. Any leading hyphens on the flag name are
// ignored.
func parse(fs *flag.FlagSet) func(string, io.Reader, func(string, string) error) error {
	return func(prefix string, r io.Reader, set func(name, value string) error) error {

		var flags []string
		fs.VisitAll(func(f *flag.Flag) {
			flags = append(flags, f.Name)
		})

		s := bufio.NewScanner(r)
		for s.Scan() {
			line := strings.TrimSpace(s.Text())
			if line == "" {
				continue // skip empties
			}

			if line[0] == '#' {
				continue // skip comments
			}

			line = strings.TrimPrefix(line, prefix+"_")

			var (
				name  string
				value string
				index = strings.IndexRune(line, '=')
			)
			if index < 0 {
				return fmt.Errorf("wrong format in env file, must be: name=value")
			}

			name, value = strings.ToLower(line[:index]), line[index+1:]

			if i := strings.Index(value, " #"); i >= 0 {
				value = strings.TrimSpace(value[:i])
			}

			var errs []error
			for _, sep := range []string{"_", "-", ".", "/"} {
				for _, f := range flags {
					replaced := strings.ReplaceAll(name, "_", sep)
					if f == replaced {
						if err := set(replaced, value); err != nil {
							errs = append(errs, err)
							continue
						}
						goto done
					}
				}
			}
			if len(errs) > 0 {
				return errs[0]
			}
		done:
		}
		return nil
	}
}

// ParserWithPrefix returns a Parser that will remove any prefix on keys in an
// .env file. For example, given prefix "MY_APP", the line `MY_APP_FOO=bar`
// in an .env file will be evaluated as name=foo, value=bar.
func ParserWithPrefix(fs *flag.FlagSet, prefix string) func(io.Reader, func(string, string) error) error {
	return func(r io.Reader, set func(name, value string) error) error {
		return parse(fs)(prefix, r, set)
	}
}
