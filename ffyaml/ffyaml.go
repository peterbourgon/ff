// Package ffyaml provides a YAML config file parser.
package ffyaml

import (
	"errors"
	"io"

	"github.com/peterbourgon/ff/v3/internal"
	"gopkg.in/yaml.v2"
)

// Parser is a helper function that uses a default ParseConfig.
func Parser(r io.Reader, set func(name, value string) error) error {
	return (&ParseConfig{}).Parse(r, set)
}

// ParseConfig collects parameters for the YAML config file parser.
type ParseConfig struct {
	// Delimiter is used when concatenating nested node keys into a flag name.
	// The default delimiter is ".".
	Delimiter string
}

// Parse a YAML document from the provided io.Reader, using the provided set
// function to set flag values. Flag names are derived from the node names and
// their key/value pairs.
func (pc *ParseConfig) Parse(r io.Reader, set func(name, value string) error) error {
	if pc.Delimiter == "" {
		pc.Delimiter = "."
	}

	var m map[string]interface{}
	if err := yaml.NewDecoder(r).Decode(&m); err != nil && !errors.Is(err, io.EOF) {
		return err
	}

	return internal.TraverseMap(m, pc.Delimiter, set)
}
