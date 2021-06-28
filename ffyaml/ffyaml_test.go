package ffyaml_test

import (
	"errors"
	"github.com/peterbourgon/ff/v3"
	"github.com/peterbourgon/ff/v3/fftest"
	"github.com/peterbourgon/ff/v3/ffyaml"
	"testing"
	"time"
)

func TestParser(t *testing.T) {
	t.Parallel()

	for _, testcase := range []struct {
		name string
		file string
		want fftest.Vars
	}{
		{
			name: "empty input",
			file: "testdata/empty_input.yaml",
			want: fftest.Vars{},
		},
		{
			name: "basic KV pairs",
			file: "testdata/basic.yaml",
			want: fftest.Vars{S: "hello", I: 10, B: true, D: 5 * time.Second, F: 3.14},
		},
		{
			name: "invalid prefix",
			file: "testdata/invalid_prefix.yaml",
			want: fftest.Vars{WantParseErrorString: "found character that cannot start any token"},
		},
		{
			name: "no value for string key",
			file: "testdata/no_value.yaml",
			want: fftest.Vars{S: "", I: 123},
		},
		{
			name: "basic arrays",
			file: "testdata/basic_array.yaml",
			want: fftest.Vars{S: "c", X: []string{"a", "b", "c"}},
		},
		{
			name: "multiline arrays",
			file: "testdata/multi_line_array.yaml",
			want: fftest.Vars{S: "c", X: []string{"d", "e", "f"}},
		},
		{
			name: "line break arrays",
			file: "testdata/line_break_array.yaml",
			want: fftest.Vars{X: []string{"first string", "second string", "third"}},
		},
		{
			name: "unquoted strings in arrays",
			file: "testdata/unquoted_string_array.yaml",
			want: fftest.Vars{X: []string{"one", "two", "three"}},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			fs, vars := fftest.Pair()
			vars.ParseError = ff.Parse(fs, []string{},
				ff.WithConfigFile(testcase.file),
				ff.WithConfigFileParser(ffyaml.Parser),
				ff.WithAllowMissingConfigFile(true),
			)
			if err := fftest.Compare(&testcase.want, vars); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestParserAllTypesEmptyVals(t *testing.T) {
	t.Parallel()

	for _, testcase := range []struct {
		name string
		file string
		want fftest.Vars
	}{
		{
			name: "basic empty vals",
			file: "testdata/empty_basic_vals.yaml",
			want: fftest.Vars{S: "", I: 0, F: 0.00, D: 0 * time.Second, B: false, X: nil},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			fs, testVars := fftest.EmptyValsFS()
			testVars.ParseError = ff.Parse(fs, []string{},
				ff.WithConfigFile(testcase.file),
				ff.WithConfigFileParser(ffyaml.Parser),
			)
			pError := testVars.ParseError
			checkParseErr(t, pError)

			if err := fftest.Compare(&testcase.want, testVars); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func checkParseErr(t *testing.T, pError error) {
	if pError != nil {
		var errMsg string
		errMsg = "ParseError: " + pError.Error()
		if errors.Is(pError, ff.StringConversionError{}) {
			errMsg = errMsg + ";errType: " + "StringConversionError"
		}
		t.Fatal(errMsg)
	}
}

func TestEmptyValsDontOverwritePresets(t *testing.T) {
	t.Parallel()

	for _, testcase := range []struct {
		name string
		file string
		want fftest.Vars
	}{
		{
			name: "preset FS vals not overwritten",
			file: "testdata/empty_basic_vals.yaml",
			want: fftest.Vars{S: "EMPTY_DEFAULT", I: -500000, F: 42.42,
				D: 86400 * time.Second, B: true, X: []string{"strVal1", "strVal2"}},
		},
	} {
		t.Run(testcase.name, func (t *testing.T){
			fs, testVars := fftest.PresetValsFS()
			testVars.ParseError = ff.Parse(fs, []string{},
				ff.WithConfigFile(testcase.file),
				ff.WithConfigFileParser(ffyaml.Parser),
			)
			pError := testVars.ParseError
			checkParseErr(t, pError)

			if err := fftest.Compare(&testcase.want, testVars); err != nil {
				t.Fatal(err)
			}
		})
	}
}
