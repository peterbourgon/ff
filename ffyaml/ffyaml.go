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
// NOTE: that this YAML parser NOW supports parsing "empty node values" (resolving to "null" values in YAML;
//  see https://yaml.org/spec/1.2/spec.html#id2786563).
// ff will therefore set the flagset value for an "empty" YAML field value to the Zero value for that golang type
func Parser(r io.Reader, set func(name, value string) error) error {
	var m map[string]interface{}
	d := yaml.NewDecoder(r)
	if err := d.Decode(&m); err != nil && err != io.EOF {
		return ParseError{err}
	}
	for key, val := range m {
		values, err := valsToStrs(val)
		if err != nil {
			return ParseError{err}
		}
		for _, ptrValue := range values {
			//Check for "empty" values in yaml
			if ptrValue != nil {
				value := *ptrValue
				if err := set(key, value); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func valsToStrs(val interface{}) ([]*string, error) {
	if vals, ok := val.([]interface{}); ok {
		ss := make([]*string, len(vals))
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
	return []*string{s}, nil

}

func valToStr(val interface{}) (*string, error) {
	rtnStr := ""
	var err error

	switch v := val.(type) {
	case byte:
		rtnStr = string([]byte{v})
	case string:
		rtnStr = v
	case bool:
		rtnStr = strconv.FormatBool(v)
	case uint64:
		rtnStr = strconv.FormatUint(v, 10)
	case int:
		rtnStr = strconv.Itoa(v)
	case int64:
		rtnStr = strconv.FormatInt(v, 10)
	case float64:
		rtnStr = strconv.FormatFloat(v, 'g', -1, 64)
	case nil:
		return nil, nil
	default:
		return nil, ff.StringConversionError{Value: val}
	}

	return &rtnStr, err
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
