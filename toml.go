package ff

import (
	"io"
	"strconv"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
)

// TOMLParser is a parser for TOML file format. Flags and their values are read
// from the key/value pairs defined in the config file
func TOMLParser(r io.Reader, set func(name, value string) error) error {
	var m map[string]interface{}
	_, err := toml.DecodeReader(r, &m)
	if err != nil {
		return errors.Wrap(err, "error parsing TOML config")
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
		return "", errors.Errorf("could not convert %q (type %T) to string", val, val)
	}
}
