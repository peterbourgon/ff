// Package fftoml provides a TOML config file paser.
package fftoml

import (
	"fmt"
	"io"
	"strconv"

	"github.com/BurntSushi/toml"
	"github.com/peterbourgon/ff"
)

// Parser is a parser for TOML file format. Flags and their values are read
// from the key/value pairs defined in the config file.
func Parser(r io.Reader, set func(name, value string) error) error {
	var m map[string]interface{}
	_, err := toml.DecodeReader(r, &m)
	if err != nil {
		return ParseError{Inner: err}
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

func join(a, b string) string {
	return a + "." + b
}
