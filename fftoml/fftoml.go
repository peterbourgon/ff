// Package fftoml provides a TOML config file paser.
package fftoml

import (
	"fmt"
	"io"
	"strconv"

	"github.com/pelletier/go-toml"
	"github.com/peterbourgon/ff"
)

// Parser is a parser for TOML file format. Flags and their values are read
// from the key/value pairs defined in the config file.
func Parser(r io.Reader, set func(name, value string) error) error {
	return New().Parse(r, set)
}

// ConfigFileParser is a parser for the TOML file format. Flags and their values
// are read from the key/value pairs defined in the config file.
// Nested tables and keys are concatenated with a delimiter to derive the
// relevant flag name.
type ConfigFileParser struct {
	delimiter string
}

// New constructs and configures a ConfigFileParser using the provided options.
func New(opts ...Option) (c ConfigFileParser) {
	c.delimiter = "."
	for _, opt := range opts {
		opt(&c)
	}
	return c
}

// Parse parses the provided io.Reader as a TOML file and uses the provided set function
// to set flag names derived from the tables names and their key/value pairs.
func (c ConfigFileParser) Parse(r io.Reader, set func(name, value string) error) error {
	tree, err := toml.LoadReader(r)
	if err != nil {
		return ParseError{Inner: err}
	}

	return parseTree(tree, "", c.delimiter, set)
}

// Option is a function which changes the behavior of the TOML config file parser.
type Option func(*ConfigFileParser)

// WithTableDelimiter is an option which configures a delimiter
// used to prefix table names onto keys when constructing
// their associated flag name.
// The default delimiter is "."
//
// For example, given the following TOML
//
//     [section.subsection]
//     value = 10
//
// Parse will match to a flag with the name `-section.subsection.value` by default.
// If the delimiter is "-", Parse will match to `-section-subsection-value` instead.
func WithTableDelimiter(d string) Option {
	return func(c *ConfigFileParser) {
		c.delimiter = d
	}
}

func parseTree(tree *toml.Tree, parent, delimiter string, set func(name, value string) error) error {
	for _, key := range tree.Keys() {
		name := key
		if parent != "" {
			name = parent + delimiter + key
		}
		switch t := tree.Get(key).(type) {
		case *toml.Tree:
			if err := parseTree(t, name, delimiter, set); err != nil {
				return err
			}
		case interface{}:
			values, err := valsToStrs(t)
			if err != nil {
				return ParseError{Inner: err}
			}
			for _, value := range values {
				if err = set(name, value); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func valsToStrs(val interface{}) ([]string, error) {
	if vals, ok := val.([]interface{}); ok {
		ss := make([]string, len(vals))
		for i := range vals {
			s, err := valToStr(vals[i])
			if err != nil {
				return nil, err
			}
			ss[i] = s
		}
		return ss, nil
	}
	s, err := valToStr(val)
	if err != nil {
		return nil, err
	}
	return []string{s}, nil

}

func valToStr(val interface{}) (string, error) {
	switch v := val.(type) {
	case string:
		return v, nil
	case bool:
		return strconv.FormatBool(v), nil
	case uint64:
		return strconv.FormatUint(v, 10), nil
	case int64:
		return strconv.FormatInt(v, 10), nil
	case float64:
		return strconv.FormatFloat(v, 'g', -1, 64), nil
	default:
		return "", ff.StringConversionError{Value: val}
	}
}

// ParseError wraps all errors originating from the TOML parser.
type ParseError struct {
	Inner error
}

// Error implenents the error interface.
func (e ParseError) Error() string {
	return fmt.Sprintf("error parsing TOML config: %v", e.Inner)
}

// Unwrap implements the xerrors.Wrapper interface, allowing
// xerrors.Is and xerrors.As to work with ParseErrors.
func (e ParseError) Unwrap() error {
	return e.Inner
}
