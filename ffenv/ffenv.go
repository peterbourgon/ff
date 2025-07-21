// Package ffenv provides an .env config file parser.
package ffenv

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Parse is a parser for .env files.
//
// Each line in the .env file is tokenized on the first `=` character. The first
// token is interpreted as the env var representation of the flag name, and the
// second token is interpreted as the value. Both tokens are trimmed of leading
// and trailing whitespace. If the value is "double quoted", control characters
// like `\n` are expanded. Lines beginning with `#` are interpreted as comments.
// End-of-line comments are not supported.
//
// Parse options related to environment variables, like [ff.WithEnvVarPrefix],
// [ff.WithEnvVarShortNames], and [ff.WithEnvVarCaseSensitive], also apply to
// .env files. For example, WithEnvVarPrefix("MYPROG") means that the keys in an
// .env file must begin with MYPROG_.
//
// If for any reason any key in an .env file matches multiple flags, parse will
// return [ff.ErrDuplicateFlag]. This can happen if you have flags with names
// that differ only in capitalization, e.g. `-v` and `-V`. If you want to
// support setting these flags from an .env file, either use discrete long
// names, or [ff.WithEnvVarCaseSensitive].
//
// Using the .env config file parser doesn't automatically enable parsing of
// actual environment variables. To do so, callers must still explicitly provide
// e.g. [ff.WithEnvVars] or [ff.WithEnvVarPrefix].
func Parse(r io.Reader, set func(name, value string) error) error {
	s := bufio.NewScanner(r)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" {
			continue // skip empties
		}

		if line[0] == '#' {
			continue // skip comments
		}

		index := strings.IndexRune(line, '=')
		if index < 0 {
			return fmt.Errorf("%w: %s", ErrInvalidLine, line)
		}

		var (
			name  = strings.TrimSpace(line[:index])
			value = strings.TrimSpace(line[index+1:])
		)

		if len(name) <= 0 {
			return fmt.Errorf("%w: %s", ErrInvalidLine, line)
		}

		if unquoted, err := strconv.Unquote(value); err == nil {
			value = unquoted
		}

		if err := set(name, value); err != nil {
			return err
		}
	}
	return nil
}

// ErrInvalidLine is returned when the parser encounters an invalid line.
var ErrInvalidLine = errors.New("invalid line")
