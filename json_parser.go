package ff

import (
	"encoding/json"
	"io"

	"github.com/peterbourgon/ff/v3/internal"
)

// JSONParser is a helper function that uses a default JSONParseConfig.
func JSONParser(r io.Reader, set func(name, value string) error) error {
	return (&ParseConfig{}).Parse(r, set)
}

// JSONParseConfig collects parameters for the JSON config file parser.
type ParseConfig struct {
	// Delimiter is used when concatenating nested node keys into a flag name.
	// The default delimiter is ".".
	Delimiter string
}

// Parse a JSON document from the provided io.Reader, using the provided set
// function to set flag values. Flag names are derived from the node names and
// their key/value pairs.
func (pc *ParseConfig) Parse(r io.Reader, set func(name, value string) error) error {
	if pc.Delimiter == "" {
		pc.Delimiter = "."
	}

	d := json.NewDecoder(r)
	d.UseNumber() // required for stringifying values

	var m map[string]interface{}
	if err := d.Decode(&m); err != nil {
		return err
	}

	return internal.TraverseMap(m, pc.Delimiter, set)
}
