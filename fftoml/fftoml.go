// Package fftoml provides a TOML config file paser.
package fftoml

import (
	"fmt"
	"io"

	"github.com/pelletier/go-toml"
	"github.com/peterbourgon/ff/v3/internal"
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
	var m map[string]any
	if err := toml.NewDecoder(r).Decode(&m); err != nil {
		return ParseError{Inner: err}
	}

	if err := internal.TraverseMap(m, c.delimiter, set); err != nil {
		return ParseError{Inner: err}
	}

	return nil
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
//	[section.subsection]
//	value = 10
//
// Parse will match to a flag with the name `-section.subsection.value` by default.
// If the delimiter is "-", Parse will match to `-section-subsection-value` instead.
func WithTableDelimiter(d string) Option {
	return func(c *ConfigFileParser) {
		c.delimiter = d
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

// Unwrap implements the errors.Wrapper interface, allowing errors.Is and
// errors.As to work with ParseErrors.
func (e ParseError) Unwrap() error {
	return e.Inner
}
