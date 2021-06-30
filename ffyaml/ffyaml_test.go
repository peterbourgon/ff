package ffyaml_test

import (
	"errors"
	"testing"
	"time"

	"github.com/peterbourgon/ff/v3"
	"github.com/peterbourgon/ff/v3/fftest"
	"github.com/peterbourgon/ff/v3/ffyaml"
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

func TestParserReturnsErrorsForBlankNonStrings(t *testing.T) {
	t.Parallel()

	for _, testcase := range []struct {
		name       string
		file       string
		wantFS     fftest.Vars
		wantErrMsg string
	}{
		{
			name:       "blank string and int vals",
			file:       "testdata/empty_str_int.yaml",
			wantErrMsg: "error setting flag \"emptyInt\" from config file: parse error",
		},
		{
			name:       "blank string and bool vals",
			file:       "testdata/empty_str_bool.yaml",
			wantErrMsg: "error setting flag \"emptyBool\" from config file: parse error",
		},
		{
			name:       "blank string and duration vals",
			file:       "testdata/empty_str_dur.yaml",
			wantErrMsg: "error setting flag \"emptyDur\" from config file: parse error",
		},
		{
			name:       "blank string and float vals",
			file:       "testdata/empty_str_float.yaml",
			wantErrMsg: "error setting flag \"emptyFloat\" from config file: parse error",
		},
		{
			name:   "blank string and slice vals",
			file:   "testdata/empty_str_slice.yaml",
			wantFS: fftest.Vars{S: "", X: []string{""}},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			fs, testVars := fftest.EmptyValsFS()
			testVars.ParseError = ff.Parse(fs, []string{},
				ff.WithConfigFile(testcase.file),
				ff.WithConfigFileParser(ffyaml.Parser),
			)
			pError := testVars.ParseError
			if pError != nil {
				var errMsg string
				errMsg = pError.Error()
				if errors.Is(pError, ff.StringConversionError{}) {
					errMsg = errMsg + ";errType: " + "StringConversionError"
				}

				if testcase.wantErrMsg != errMsg {
					t.Fatal("error \"" + errMsg + "\" Does Not Match Expected ErrorMsg: " + testcase.wantErrMsg)
				}
			}

			if err := fftest.Compare(&testcase.wantFS, testVars); err != nil {
				t.Fatal(err)
			}
		})
	}
}
