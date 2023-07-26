package ffyaml_test

import (
	"os"
	"testing"
	"time"

	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/fftest"
	"github.com/peterbourgon/ff/v4/ffyaml"
)

func TestParser(t *testing.T) {
	t.Parallel()

	testcases := fftest.TestCases{
		{
			Name:       "empty",
			ConfigFile: "testdata/empty.yaml",
			Want:       fftest.Vars{},
		},
		{
			Name:       "basic KV pairs",
			ConfigFile: "testdata/basic.yaml",
			Want:       fftest.Vars{S: "hello", I: 10, B: true, D: 5 * time.Second, F: 3.14},
		},
		{
			Name:       "invalid prefix",
			ConfigFile: "testdata/invalid_prefix.yaml",
			Want:       fftest.Vars{WantParseErrorString: "found character that cannot start any token"},
		},
		{
			Name:       "no value for s",
			Default:    fftest.Vars{S: "xxx", I: 123, F: 9.99},
			ConfigFile: "testdata/no_value_s.yaml",
			Want:       fftest.Vars{S: "", I: 123, F: 9.99},
		},
		{
			Name:       "no value for i",
			Default:    fftest.Vars{S: "xxx", I: 123, F: 9.99},
			ConfigFile: "testdata/no_value_i.yaml",
			Want:       fftest.Vars{WantParseErrorString: "parse error"},
		},
		{
			Name:       "basic arrays",
			ConfigFile: "testdata/basic_array.yaml",
			Want:       fftest.Vars{S: "c", X: []string{"a", "b", "c"}},
		},
		{
			Name:       "multiline arrays",
			ConfigFile: "testdata/multi_line_array.yaml",
			Want:       fftest.Vars{S: "c", X: []string{"d", "e", "f"}},
		},
		{
			Name:       "line break arrays",
			ConfigFile: "testdata/line_break_array.yaml",
			Want:       fftest.Vars{X: []string{"first string", "second string", "third"}},
		},
		{
			Name:       "unquoted strings in arrays",
			ConfigFile: "testdata/unquoted_string_array.yaml",
			Want:       fftest.Vars{X: []string{"one", "two", "three"}},
		},
		{
			Name:       "missing config file allowed",
			ConfigFile: "testdata/this_file_does_not_exist.yaml",
			Options:    []ff.Option{ff.WithConfigAllowMissingFile()},
			Want:       fftest.Vars{},
		},
		{
			Name:       "missing config file not allowed",
			ConfigFile: "testdata/this_file_does_not_exist.yaml",
			Want:       fftest.Vars{WantParseErrorIs: os.ErrNotExist},
		},
		{
			Name:         "nested with '.'",
			ConfigFile:   "testdata/nested.yaml",
			Constructors: []fftest.Constructor{fftest.NewNestedConstructor(".")},
			Want:         fftest.Vars{S: "a string", F: 1.23, B: true, X: []string{"one", "two", "three"}},
		},
		{
			Name:         "nested with '-'",
			ConfigFile:   "testdata/nested.yaml",
			Default:      fftest.Vars{A: true},
			Constructors: []fftest.Constructor{fftest.NewNestedConstructor("-")},
			Options:      []ff.Option{ff.WithConfigFileParser(ffyaml.Parser{Delimiter: "-"}.Parse)},
			Want:         fftest.Vars{S: "a string", F: 1.23, A: true, B: true, X: []string{"one", "two", "three"}},
		},
	}

	testcases.Run(t, ff.WithConfigFileParser(ffyaml.Parse))
}
