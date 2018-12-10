package ff

import (
	"encoding/json"
	"io"
	"strconv"

	"github.com/pkg/errors"
)

// JSONParser is a parser for config files in JSON format. Input should be
// an object. The object's keys are treated as flag names, and the object's
// values as flag values. If the value is an array, the flag will be set
// multiple times.
func JSONParser(r io.Reader, set func(name, value string) error) error {
	var m map[string]interface{}
	d := json.NewDecoder(r)
	// Must set UseNumber for stringifyValue to work
	d.UseNumber()
	err := d.Decode(&m)
	if err != nil {
		return errors.Wrap(err, "error parsing JSON config")
	}
	for key, val := range m {
		values, err := stringifySlice(val)
		if err != nil {
			return errors.Wrap(err, "error parsing JSON config")
		}
		for _, value := range values {
			if err := set(key, value); err != nil {
				return err
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
		return "", errors.Errorf("could not convert %q (type %T) to string", val, val)
	}
}
