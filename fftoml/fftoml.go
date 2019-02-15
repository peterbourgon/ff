package fftoml

import (
	"fmt"
	"io"
	"strconv"

	"github.com/BurntSushi/toml"
)

// Parser is a parser for TOML file format. Flags and their values are read
// from the key/value pairs defined in the config file
func Parser(r io.Reader, set func(name, value string) error) error {
	var m map[string]interface{}
	_, err := toml.DecodeReader(r, &m)
	if err != nil {
		return fmt.Errorf("error parsing TOML config: %v", err)
	}
	for key, val := range m {
		value, err := valToStr(val)
		if err != nil {
			return err
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
		return "", fmt.Errorf("could not convert %q (type %T) to string", val, val)
	}
}
