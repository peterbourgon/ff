package ff

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
)

// JSONParser is a parser for config files in JSON format. Input should be
// an object. The object's keys are treated as flag names, and the object's
// values as flag values. If the value is an array, the flag will be set
// multiple times.
func JSONParser(r io.Reader, set func(name, value string) error) error {
	return NewJSONParser().Parse(r, set)
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
	return parseObject(m, "", c.delimiter, set)
}

// JSONOption changes the behavior of the JSON config file parser.
type JSONOption func(*JSONConfigFileParser)

// WithObjectDelimiter is an option which configures a delimiter
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
func WithObjectDelimiter(d string) JSONOption {
	return func(p *JSONConfigFileParser) {
		p.delimiter = d
	}
}

func parseObject(obj map[string]interface{}, parent, delimiter string, set func(name, value string) error) error {
	for key, val := range obj {
		name := key
		if parent != "" {
			name = parent + delimiter + key
		}
		switch n := val.(type) {
		case map[string]interface{}:
			if err := parseObject(n, name, delimiter, set); err != nil {
				return err
			}
		default:
			values, err := stringifySlice(val)
			if err != nil {
				return JSONParseError{Inner: err}
			}
			for _, value := range values {
				if err := set(name, value); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func stringifySlice(val interface{}) ([]string, error) {
	if vals, ok := val.([]interface{}); ok {
		ss := make([]string, len(vals))
		for i := range vals {
			s, err := stringifyValue(vals[i])
			if err != nil {
				return nil, err
			}
			ss[i] = s
		}
		return ss, nil
	}
	s, err := stringifyValue(val)
	if err != nil {
		return nil, err
	}
	return []string{s}, nil
}

func stringifyValue(val interface{}) (string, error) {
	switch v := val.(type) {
	case string:
		return v, nil
	case json.Number:
		return v.String(), nil
	case bool:
		return strconv.FormatBool(v), nil
	default:
		return "", StringConversionError{Value: val}
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
