package ff

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/peterbourgon/ff/v3/internal"
)

// JSONParser is a parser for config files in JSON format. Input should be
// an object. The object's keys are treated as flag names, and the object's
// values as flag values. If the value is an array, the flag will be set
// multiple times.
func JSONParser(r io.Reader, set func(name, value string) error) error {
	return NewJSONConfigFileParser().Parse(r, set)
}

// JSONConfigFileParser is a parser for the JSON file format. Flags and their values
// are read from the key/value pairs defined in the config file.
// Nested objects and keys are concatenated with a delimiter to derive the
// relevant flag name.
type JSONConfigFileParser struct {
	delimiter string
}

// NewJSONConfigFileParser returns a JSONConfigFileParser using the provided options.
func NewJSONConfigFileParser(opts ...JSONOption) *JSONConfigFileParser {
	p := &JSONConfigFileParser{
		delimiter: ".",
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// Parse a JSON object from the provided io.Reader, using the provided set function
// to set flag names derived from the node names and their key/value pairs.
func (p *JSONConfigFileParser) Parse(r io.Reader, set func(name, value string) error) error {
	var m map[string]interface{}
	d := json.NewDecoder(r)
	d.UseNumber() // must set UseNumber for stringifyValue to work
	if err := d.Decode(&m); err != nil {
		return JSONParseError{Inner: err}
	}

	if err := internal.TraverseMap(m, p.delimiter, set); err != nil {
		return StringConversionError{Value: err}
	}
	return nil
}

// JSONOption changes the behavior of the JSON config file parser.
type JSONOption func(*JSONConfigFileParser)

// WithJSONDelimiter is an option which configures a delimiter
// used to prefix object names onto keys when constructing
// their associated flag name.
// The default delimiter is "."
//
// For example, given the following JSON
//
//	{
//		"section": {
//			"subsection": {
//				"value": 10
//			}
//		}
//	}
//
// Parse will match to a flag with the name `-section.subsection.value` by default.
// If the delimiter is "-", Parse will match to `-section-subsection-value` instead.
func WithJSONDelimiter(d string) JSONOption {
	return func(p *JSONConfigFileParser) {
		p.delimiter = d
	}
}

// JSONParseError wraps all errors originating from the JSONParser.
type JSONParseError struct {
	Inner error
}

// Error implenents the error interface.
func (e JSONParseError) Error() string {
	return fmt.Sprintf("error parsing JSON config: %v", e.Inner)
}

// Unwrap implements the errors.Wrapper interface, allowing errors.Is and
// errors.As to work with JSONParseErrors.
func (e JSONParseError) Unwrap() error {
	return e.Inner
}

// StringConversionError is returned when a value in a config file
// can't be converted to a string, to be provided to a flag.
type StringConversionError struct {
	Value interface{}
}

// Error implements the error interface.
func (e StringConversionError) Error() string {
	return fmt.Sprintf("couldn't convert %q (type %T) to string", e.Value, e.Value)
}
