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

// Parse is a parser for .env files. Each line is tokenized on the first `=`
// character. The first token is interpreted as the env var representation of
// the flag name, and the second token is interpreted as the value. Both tokens
// are trimmed of leading and trailing whitespace. If the value is "double
// quoted", control characters like `\n` are expanded. Lines beginning with `#`
// are interpreted as comments. End-of-line comments are not supported.
//
// The parser respects the [ff.WithEnvVarPrefix] option. For example, if parse
// is called with an env var prefix MYPROG, then both FOO=bar and MYPROG_FOO=bar
// would set a flag named foo.
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

		if len(value) <= 0 {
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
