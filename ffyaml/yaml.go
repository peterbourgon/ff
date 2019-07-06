package ffyaml

import (
	"fmt"
	"io"

	"gopkg.in/yaml.v2"
)

func Parser(r io.Reader, set func(name, value string) error) error {
	var m map[string]string
	d := yaml.NewDecoder(r)
	if err := d.Decode(&m); err != nil && err != io.EOF {
		return fmt.Errorf("error parsing YAML config: %v", err)
	}
	for key, val := range m {
		if err := set(key, val); err != nil {
			return err
		}
	}
	return nil
}
