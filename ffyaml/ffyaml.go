// Package ffyaml provides a YAML config file parser.
package ffyaml

import (
	"fmt"
	"io"
	"strconv"

	"github.com/peterbourgon/ff/v3"
	"gopkg.in/yaml.v3"
)

// Parser is a parser for YAML file format. Flags and their values are read
// from the key/value pairs defined in the config file.
func Parser(r io.Reader, set func(name, value string) error) error {
	return New().Parse(r, set)
}

// ConfigFileParser is a parser for the YAML file format. Flags and their
// values are read from the key/value pairs defined in the config file. The
// parser can be pointed to a section of the file via a key path. This allows
// one to: place different configurations in the same file, use same file for
// other tools, or use the same file for different instances of this parser.
type ConfigFileParser struct {
	path []string
}

// New constructs and configures a ConfigFileParser using the provided options.
func New(opts ...Option) (c ConfigFileParser) {
	for _, opt := range opts {
		opt(&c)
	}
	return c
}

// Parse parses the provided io.Reader as a YAML file and uses the set function
// to set flag names derived from their key/value pairs.
func (c ConfigFileParser) Parse(r io.Reader, set func(name, value string) error) error {
	var m map[string]interface{}
	d := yaml.NewDecoder(r)
	if err := d.Decode(&m); err != nil && err != io.EOF {
		return ParseError{err}
	}
	for i, key := range c.path {
		val, ok := m[key]
		if !ok {
			return ParseError{fmt.Errorf("key path '%s' not found", c.path[0:i+1])}
		}
		m, ok = val.(map[string]interface{})
		if !ok {
			return ParseError{fmt.Errorf("key path '%s' not a YAML map", c.path[0:i+1])}
		}
	}
	for key, val := range m {
		values, err := valsToStrs(val)
		if err != nil {
			return ParseError{err}
		}
		for _, value := range values {
			if err := set(key, value); err != nil {
				return err
			}
		}
	}
	return nil
}

// Option is a function which changes the behavior of the YAML config file parser.
type Option func(*ConfigFileParser)

// WithKeyPath is an option that specifies a path of keys leading to the section
// of the YAML config used by the configuration parser. The default path is nil,
// which means the root configuration is used.
//
// For example, given the following YAML
//
//	config:
//	  dev:
//	    value: 10
//	  prod:
//	    value: 100
//
// Specifying a path []string{"config", "dev"} will yield a value of 10 for value.
func WithKeyPath(path ...string) Option {
	return func(c *ConfigFileParser) {
		c.path = path
	}
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
	case byte:
		return string([]byte{v}), nil
	case string:
		return v, nil
	case bool:
		return strconv.FormatBool(v), nil
	case uint64:
		return strconv.FormatUint(v, 10), nil
	case int:
		return strconv.Itoa(v), nil
	case int64:
		return strconv.FormatInt(v, 10), nil
	case float64:
		return strconv.FormatFloat(v, 'g', -1, 64), nil
	case nil:
		return "", nil
	default:
		return "", ff.StringConversionError{Value: val}
	}
}

// ParseError wraps all errors originating from the YAML parser.
type ParseError struct {
	Inner error
}

// Error implenents the error interface.
func (e ParseError) Error() string {
	return fmt.Sprintf("error parsing YAML config: %v", e.Inner)
}

// Unwrap implements the errors.Wrapper interface, allowing errors.Is and
// errors.As to work with ParseErrors.
func (e ParseError) Unwrap() error {
	return e.Inner
}
