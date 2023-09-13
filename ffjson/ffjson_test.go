package ffjson_test

import (
	"io"
	"testing"
	"time"

	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffjson"
	"github.com/peterbourgon/ff/v4/fftest"
)

func TestParser(t *testing.T) {
	t.Parallel()

	testcases := fftest.TestCases{
		{
			Name:       "empty input",
			ConfigFile: "testdata/empty.json",
			Want:       fftest.Vars{},
		},
		{
			Name:       "basic KV pairs",
			ConfigFile: "testdata/basic.json",
			Want:       fftest.Vars{S: "s", I: 10, B: true, D: 5 * time.Second},
		},
		{
			Name:       "value arrays",
			ConfigFile: "testdata/value_arrays.json",
			Want:       fftest.Vars{S: "bb", I: 12, B: true, D: 5 * time.Second, X: []string{"a", "B", "👍"}},
		},
		{
			Name:       "bad JSON file",
			ConfigFile: "testdata/bad.json",
			Want:       fftest.Vars{WantParseErrorIs: io.ErrUnexpectedEOF},
		},
		{
			Name:         "nested with '.'",
			ConfigFile:   "testdata/nested.json",
			Constructors: []fftest.Constructor{fftest.NewNestedConstructor(".")},
			Want:         fftest.Vars{S: "foo bar", I: 34, X: []string{"alpha", "beta", "delta"}},
		},
		{
			Name:         "nested with '-'",
			ConfigFile:   "testdata/nested.json",
			Constructors: []fftest.Constructor{fftest.NewNestedConstructor("-")},
			Options:      []ff.Option{ff.WithConfigFileParser(ffjson.Parser{Delimiter: "-"}.Parse)},
			Want:         fftest.Vars{S: "foo bar", I: 34, X: []string{"alpha", "beta", "delta"}},
		},
	}

	testcases.Run(t)
}
