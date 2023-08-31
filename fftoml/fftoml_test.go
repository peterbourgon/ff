package fftoml_test

import (
	"testing"
	"time"

	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/fftest"
	"github.com/peterbourgon/ff/v4/fftoml"
)

func TestParser(t *testing.T) {
	t.Parallel()

	testcases := fftest.TestCases{
		{
			Name:       "empty input",
			ConfigFile: "testdata/empty.toml",
			Want:       fftest.Vars{},
		},
		{
			Name:       "basic KV pairs",
			ConfigFile: "testdata/basic.toml",
			Want: fftest.Vars{
				S: "s",
				I: 10,
				F: 3.14e10,
				B: true,
				D: 5 * time.Second,
				X: []string{"1", "a", "👍"},
			},
		},
		{
			Name:       "bad TOML file",
			ConfigFile: "testdata/bad.toml",
			Want:       fftest.Vars{WantParseErrorString: "invalid character at start of key"},
		},
		{
			Name:         "nested with '.'",
			ConfigFile:   "testdata/table.toml",
			Default:      fftest.Vars{I: 999},
			Constructors: []fftest.Constructor{fftest.NewNestedConstructor(".")},
			Want:         fftest.Vars{S: "a string", I: 999, F: 1.23, X: []string{"one", "two", "three"}},
		},
		{
			Name:         "nested with '-'",
			ConfigFile:   "testdata/table.toml",
			Constructors: []fftest.Constructor{fftest.NewNestedConstructor("-")},
			Options:      []ff.Option{ff.WithConfigFileParser(fftoml.Parser{Delimiter: "-"}.Parse)},
			Want:         fftest.Vars{S: "a string", F: 1.23, X: []string{"one", "two", "three"}},
		},
	}

	testcases.Run(t, ff.WithConfigFileParser(fftoml.Parse))
}
