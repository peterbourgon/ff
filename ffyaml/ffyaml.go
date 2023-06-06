// Package ffyaml provides a YAML config file parser.
package ffyaml

import (
	"fmt"
	"io"

	"github.com/peterbourgon/ff/v3"
	"github.com/peterbourgon/ff/v3/internal"
	"gopkg.in/yaml.v2"
)

// Parser is a parser for YAML file format. Flags and their values are read
// from the key/value pairs defined in the config file.
func Parser(r io.Reader, set func(name, value string) error) error {
	return NewConfigFileParser().Parse(r, set)
}

// ConfigFileParser is a parser for the YAML file format. Flags and their values
// are read from the key/value pairs defined in the config file.
// Nested nodes and keys are concatenated with a delimiter to derive the
// relevant flag name.
type ConfigFileParser struct {
	delimiter string
}

// NewConfigFileParser returns a ConfigFileParser with the provided options.
func NewConfigFileParser(opts ...Option) *ConfigFileParser {
	p := &ConfigFileParser{
		delimiter: ".",
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// Parse a YAML document from the provided io.Reader, using the provided set function
// to set flag names derived from the node names and their key/value pairs.
func (p *ConfigFileParser) Parse(r io.Reader, set func(name, value string) error) error {
	var m map[string]interface{}
	d := yaml.NewDecoder(r)
	if err := d.Decode(&m); err != nil && err != io.EOF {
		return ParseError{err}
	}

	if err := internal.TraverseMap(m, p.delimiter, set); err != nil {
		return ff.StringConversionError{Value: err}
	}
	return nil
}

// Option changes the behavior of the YAML config file parser.
type Option func(*ConfigFileParser)

// WithDelimiter is an option which configures a delimiter
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
func WithDelimiter(d string) Option {
	return func(c *ConfigFileParser) {
		c.delimiter = d
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
