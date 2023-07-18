// Package fftoml provides a TOML config file paser.
package fftoml

import (
	"io"

	"github.com/pelletier/go-toml"
	"github.com/peterbourgon/ff/v3/internal"
)

// Parser is a helper function that uses a default ParseConfig.
func Parser(r io.Reader, set func(name, value string) error) error {
	return (&ParseConfig{}).Parse(r, set)
}

// ParseConfig collects parameters for the TOML config file parser.
type ParseConfig struct {
	// Delimiter is used when concatenating nested node keys into a flag name.
	// The default delimiter is ".".
	Delimiter string
}

// Parse a TOML document from the provided io.Reader, using the provided set
// function to set flag values. Flag names are derived from the node names and
// their key/value pairs.
func (pc *ParseConfig) Parse(r io.Reader, set func(name, value string) error) error {
	if pc.Delimiter == "" {
		pc.Delimiter = "."
	}

	var m map[string]any
	if err := toml.NewDecoder(r).Decode(&m); err != nil {
		return err
	}

	return internal.TraverseMap(m, pc.Delimiter, set)
}
