package ff

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// SourceType enumerates the sources from which we can retrieve values.
type SourceType uint8

const (
	// SourceTypeEnv takes variables from the environment.
	SourceTypeEnv SourceType = iota

	// SourceTypeJSON takes variables from a JSON object serialized to a file.
	SourceTypeJSON
)

// Source from which variable values may be retrieved.
type Source interface {
	Type() SourceType
	Fetch(*FlagSet) map[string]string
}

type source struct {
	t SourceType
	f func(*FlagSet) map[string]string
}

func (src source) Type() SourceType                    { return src.t }
func (src source) Fetch(fs *FlagSet) map[string]string { return src.f(fs) }

// FromEnvironment returns a source that takes values from environment variables
// with the given prefix. The prefix is not stripped from the selected vars.
func FromEnvironment(prefix string) Source {
	return source{SourceTypeEnv, func(*FlagSet) map[string]string {
		m := map[string]string{}
		for _, kv := range os.Environ() {
			toks := strings.SplitN(kv, "=", 2)
			k, v := toks[0], toks[1]
			if strings.HasPrefix(k, prefix) {
				m[k] = v
			}
		}
		return m
	}}
}

// FromJSONFileVia returns a parse source that takes values from a JSON object
// encoded in a file referenced by the passed variable name. The file must
// contain a single JSON object, with keys at the top level, and values as
// strings.
func FromJSONFileVia(configFileFlagName string) Source {
	return source{SourceTypeJSON, func(fs *FlagSet) map[string]string {
		v, ok := fs.vars[configFileFlagName]
		if !ok {
			panic(fmt.Sprintf("JSON config file flag name %q not defined", configFileFlagName))
		}
		filename := v.value.String()
		return FromJSONFile(filename).Fetch(fs)
	}}
}

// FromJSONFile returns a source that takes values from a JSON object encoded in
// the file. The file must contain a single JSON object, with keys at the top
// level, and values as strings.
func FromJSONFile(filename string) Source {
	return source{SourceTypeJSON, func(fs *FlagSet) map[string]string {
		if filename == "" {
			return map[string]string{}
		}
		f, err := os.Open(filename)
		if err != nil {
			panic(fmt.Sprintf("JSON file couldn't be opened: %v", err))
		}
		defer f.Close()
		return FromJSONReader(f).Fetch(fs)
	}}
}

// FromJSONReader returns a parse source that unmarshals a map[string]string
// from the provided reader. The reader must contain a single JSON object, with
// keys at the top level, and values as strings.
func FromJSONReader(r io.Reader) Source {
	return source{SourceTypeJSON, func(fs *FlagSet) map[string]string {
		var m map[string]string
		if err := json.NewDecoder(r).Decode(&m); err != nil {
			panic(fmt.Sprintf("JSON couldn't be decoded: %v", err))
		}
		return m
	}}
}
