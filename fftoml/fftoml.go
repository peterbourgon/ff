package fftoml

import (
	"fmt"
	"io"
	"strconv"

	"github.com/BurntSushi/toml"
	"github.com/peterbourgon/ff"
)

// Parser is a parser for TOML file format. Flags and their values are read
// from the key/value pairs defined in the config file
func Parser(r io.Reader, set func(name, value string) error) error {
	var m map[string]interface{}
	_, err := toml.DecodeReader(r, &m)
	if err != nil {
		return ParseError{Inner: err}
	}
	for key, val := range m {
		value, err := valToStr(val)
		if err != nil {
			return ParseError{Inner: err}
		}
		if err = set(key, value); err != nil {
			return err
		}
	}
	return nil
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
