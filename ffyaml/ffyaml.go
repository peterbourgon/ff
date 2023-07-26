// Package ffyaml provides a YAML config file parser.
package ffyaml

import (
	"errors"
	"io"

	"github.com/peterbourgon/ff/v4/internal"
	"gopkg.in/yaml.v2"
)

// Parse is a helper function that uses a default parser.
func Parse(r io.Reader, set func(name, value string) error) error {
	return (&Parser{}).Parse(r, set)
}

// Parser collects parameters for the YAML config file parser.
type Parser struct {
	// Delimiter is used when concatenating nested node keys into a flag name.
	// The default delimiter is ".".
	Delimiter string
}

// Parse a YAML document from the provided io.Reader, using the provided set
// function to set flag values. Flag names are derived from the node names and
// their key/value pairs.
func (p Parser) Parse(r io.Reader, set func(name, value string) error) error {
	if p.Delimiter == "" {
		p.Delimiter = "."
	}

	var m map[string]interface{}
	if err := yaml.NewDecoder(r).Decode(&m); err != nil && !errors.Is(err, io.EOF) {
		return err
	}

	return internal.TraverseMap(m, p.Delimiter, set)
}
