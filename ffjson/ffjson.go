// Package ffjson provides a JSON config file parser.
package ffjson

import (
	"encoding/json"
	"io"

	"github.com/peterbourgon/ff/v4/internal/ffdata"
)

// Parse is a helper function that uses a default parser.
func Parse(r io.Reader, set func(name, value string) error) error {
	return (&Parser{}).Parse(r, set)
}

// Parser collects parameters for the JSON config file parser.
type Parser struct {
	// Delimiter is used when concatenating nested node keys into a flag name.
	// The default delimiter is ".".
	Delimiter string
}

// Parse a JSON document from the provided io.Reader, using the provided set
// function to set flag values. Flag names are derived from the node names and
// their key/value pairs.
func (p Parser) Parse(r io.Reader, set func(name, value string) error) error {
	if p.Delimiter == "" {
		p.Delimiter = "."
	}

	d := json.NewDecoder(r)
	d.UseNumber() // required for stringifying values

	var m map[string]interface{}
	if err := d.Decode(&m); err != nil {
		return err
	}

	return ffdata.TraverseMap(m, p.Delimiter, set)
}
