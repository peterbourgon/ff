// Package ffyaml provides a YAML config file paser.
package ffyaml

import (
	"fmt"
	"io"
	"strconv"

	"github.com/peterbourgon/ff"
	"gopkg.in/yaml.v2"
)

// Parser is a parser for YAML file format. Flags and their values are read
// from the key/value pairs defined in the config file.
func Parser(r io.Reader, set func(name, value string) error) error {
	var m map[string]interface{}
	d := yaml.NewDecoder(r)
	if err := d.Decode(&m); err != nil && err != io.EOF {
		return ParseError{err}
	}
	for key, val := range m {
		err := setup(key, val, set)
		if err != nil {
			return ParseError{err}
		}
	}
	return nil
}

func setup(key string, val interface{}, set func(name, value string) error) error {
	if obj, ok := val.(map[string]interface{}); ok {
		for sub, val := range obj {
			err := setup(join(key, sub), val, set)
			if err != nil {
				return err
			}
		}
		return nil
	}
	if vals, ok := val.([]interface{}); ok {
		for _, val := range vals {
			err := setup(key, val, set)
			if err != nil {
				return err
			}
		}
		return nil
	}
	s, err := valToStr(val)
	if err != nil {
		return err
	}
	return set(key, s)
}

func valToStr(val interface{}) (string, error) {
	switch v := val.(type) {
	case byte:
		return string(v), nil
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

// Unwrap implements the xerrors.Wrapper interface, allowing
// xerrors.Is and xerrors.As to work with ParseErrors.
func (e ParseError) Unwrap() error {
	return e.Inner
}

func join(a, b string) string {
	return a + "." + b
}
