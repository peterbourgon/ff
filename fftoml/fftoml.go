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
	return ParserWith()(r, set)
}

type config struct {
	separator string
}

// Option is a function which configures a fftoml.ConfigFileParser
type Option func(*config)

// FlagSeparator is an option which configures a separator
// to use when constructing a flag name
func FlagSeparator(s string) Option {
	return func(c *config) {
		c.separator = s
	}
}

// ParserWith configures and returns a new ConfigFileParser using the
// provided slice of Option types
// By default the returned Parser uses a "." as a separator
func ParserWith(opts ...Option) ff.ConfigFileParser {
	return func(r io.Reader, set func(name, value string) error) error {
		config := config{separator: "."}
		for _, opt := range opts {
			opt(&config)
		}
		tree, err := toml.LoadReader(r)
		if err != nil {
			return ParseError{Inner: err}
		}
		return parseTree(tree, "", config.separator, set)
	}
}

func parseTree(tree *toml.Tree, parent, separator string, set func(name, value string) error) error {
	for _, key := range tree.Keys() {
		name := key
		if parent != "" {
			name = parent + separator + key
		}
		switch t := tree.Get(key).(type) {
		case *toml.Tree:
			if err := parseTree(t, name, separator, set); err != nil {
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
