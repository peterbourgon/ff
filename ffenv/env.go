package ffenv

import (
	"bufio"
	"io"
	"strings"
)

// Parser is a parser for .env file format: flag=value. Each
// line is tokenized as a single key/value pair.
func Parser(r io.Reader, set func(name, value string) error) error {
	return parse("", r, set)
}

func parse(prefix string, r io.Reader, set func(name, value string) error) error {
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
			name, value = line, "true" // boolean option
		} else {
			name, value = strings.ToLower(line[:index]), line[index+1:]
			name = strings.ReplaceAll(name, "_", "-")
		}

		if i := strings.Index(value, " #"); i >= 0 {
			value = strings.TrimSpace(value[:i])
		}

		if err := set(name, value); err != nil {
			return err
		}
	}
	return nil
}

// ParserWithPrefix returns a Parser that will remove any prefix on keys in an
// .env file. For example, given prefix "MY_APP", the line `MY_APP_FOO=bar` 
// in an .env file will be evaluated as name=foo, value=bar.
func ParserWithPrefix(prefix string) func(io.Reader, func(string, string) error) error {
	return func(r io.Reader, set func(name, value string) error) error {
		return parse(prefix, r, set)
	}
}
