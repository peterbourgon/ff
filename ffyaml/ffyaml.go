// Package ffyaml provides a YAML config file parser.
package ffyaml

import (
	"fmt"
	"io"
	"strconv"

	"github.com/peterbourgon/ff/v3"
	"gopkg.in/yaml.v2"
)

// Parser is a parser for YAML file format. Flags and their values are read
// from the key/value pairs defined in the config file.
func Parser(r io.Reader, set func(name, value string) error) error {
	return New().Parse(r, set)
}

// ConfigFileParser is a parser for the YAML file format. Flags and their values
// are read from the key/value pairs defined in the config file.
// Nested nodes and keys are concatenated with a delimiter to derive the
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

// Parse parses the provided io.Reader as a YAML file and uses the provided set function
// to set flag names derived from the node names and their key/value pairs.
func (c ConfigFileParser) Parse(r io.Reader, set func(name, value string) error) error {
	var m map[string]interface{}
	d := yaml.NewDecoder(r)
	if err := d.Decode(&m); err != nil && err != io.EOF {
		return ParseError{err}
	}
	return parseNode(m, "", c.delimiter, set)
}

// Option is a function which changes the behavior of the YAML config file parser.
type Option func(*ConfigFileParser)

// WithNodeDelimiter is an option which configures a delimiter
// used to prefix node names onto keys when constructing
// their associated flag name.
// The default delimiter is "."
//
// For example, given the following YAML
//
//	section:
//		subsection:
//			value: 10
//
// Parse will match to a flag with the name `-section.subsection.value` by default.
// If the delimiter is "-", Parse will match to `-section-subsection-value` instead.
func WithNodeDelimiter(d string) Option {
	return func(c *ConfigFileParser) {
		c.delimiter = d
	}
}

func parseNode(node map[string]interface{}, parent, delimiter string, set func(name, value string) error) error {
	for key, val := range node {
		name := key
		if parent != "" {
			name = parent + delimiter + key
		}
		switch n := val.(type) {
		case map[interface{}]interface{}:
			m := make(map[string]interface{})
			for k, v := range n {
				m[fmt.Sprint(k)] = v
			}
			if err := parseNode(m, name, delimiter, set); err != nil {
				return err
			}
		default:
			values, err := valsToStrs(n)
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
