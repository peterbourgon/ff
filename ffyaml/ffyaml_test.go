package ffyaml_test

import (
	"flag"
	"os"
	"testing"
	"time"

	"github.com/peterbourgon/ff/v3"
	"github.com/peterbourgon/ff/v3/fftest"
	"github.com/peterbourgon/ff/v3/ffyaml"
)

func TestParser(t *testing.T) {
	t.Parallel()

	for _, testcase := range []struct {
		vars func(*flag.FlagSet) *fftest.Vars
		name string
		file string
		opts []ffyaml.Option
		miss bool // AllowMissingConfigFiles
		want fftest.Vars
	}{
		{
			name: "empty",
			file: "testdata/empty.yaml",
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
			vars: fftest.NonzeroDefaultVars,
			name: "no value for s",
			file: "testdata/no_value_s.yaml",
			want: fftest.Vars{S: "", I: 123, F: 9.99, B: true, D: 3 * time.Hour},
		},
		{
			vars: fftest.NonzeroDefaultVars,
			name: "no value for i",
			file: "testdata/no_value_i.yaml",
			want: fftest.Vars{WantParseErrorString: "parse error"},
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
		{
			name: "missing config file allowed",
			file: "testdata/this_file_does_not_exist.yaml",
			miss: true,
			want: fftest.Vars{},
		},
		{
			name: "missing config file not allowed",
			file: "testdata/this_file_does_not_exist.yaml",
			miss: false,
			want: fftest.Vars{WantParseErrorIs: os.ErrNotExist},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			if testcase.vars == nil {
				testcase.vars = fftest.DefaultVars
			}
			fs := flag.NewFlagSet("fftest", flag.ContinueOnError)
			vars := testcase.vars(fs)
			vars.ParseError = ff.Parse(fs, []string{},
				ff.WithConfigFile(testcase.file),
				ff.WithConfigFileParser(ffyaml.Parser),
				ff.WithAllowMissingConfigFile(testcase.miss),
			)
			fftest.Compare(t, &testcase.want, vars)
		})
	}
}

func TestParserWithKeyPath(t *testing.T) {
	t.Parallel()

	for _, testcase := range []struct {
		vars func(*flag.FlagSet) *fftest.Vars
		name string
		file string
		opts []ffyaml.Option
		miss bool // AllowMissingConfigFiles
		want fftest.Vars
	}{
		{
			name: "key path",
			file: "testdata/key_path.yaml",
			opts: []ffyaml.Option{ffyaml.WithKeyPath("root", "child1")},
			want: fftest.Vars{S: "child1 value"},
		},
		{
			name: "another key path",
			file: "testdata/key_path.yaml",
			opts: []ffyaml.Option{ffyaml.WithKeyPath("root", "child2")},
			want: fftest.Vars{S: "child2 value"},
		},
		{
			name: "no key path without a terminal map",
			file: "testdata/key_path.yaml",
			opts: []ffyaml.Option{ffyaml.WithKeyPath()},
			want: fftest.Vars{WantParseErrorString: "error parsing YAML config: couldn't convert map"},
		},
		{
			name: "invalid prefix",
			file: "testdata/invalid_prefix.yaml",
			want: fftest.Vars{WantParseErrorString: "found character that cannot start any token"},
		},
		{
			name: "no such key path",
			file: "testdata/key_path.yaml",
			opts: []ffyaml.Option{ffyaml.WithKeyPath("root", "child3")},
			want: fftest.Vars{WantParseErrorString: "error parsing YAML config: key path '[root child3]' not found"},
		},
		{
			name: "key path does not point to a terminal map",
			file: "testdata/key_path.yaml",
			opts: []ffyaml.Option{ffyaml.WithKeyPath("root", "child2", "s")},
			want: fftest.Vars{WantParseErrorString: "not a YAML map"},
		},
		{
			name: "key path does not terminate at a terminal map",
			file: "testdata/key_path.yaml",
			opts: []ffyaml.Option{ffyaml.WithKeyPath("root")},
			want: fftest.Vars{WantParseErrorString: "error parsing YAML config: couldn't convert map"},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			if testcase.vars == nil {
				testcase.vars = fftest.DefaultVars
			}
			fs := flag.NewFlagSet("fftest", flag.ContinueOnError)
			vars := testcase.vars(fs)
			vars.ParseError = ff.Parse(fs, []string{},
				ff.WithConfigFile(testcase.file),
				ff.WithConfigFileParser(ffyaml.New(testcase.opts...).Parse),
				ff.WithAllowMissingConfigFile(testcase.miss),
			)
			fftest.Compare(t, &testcase.want, vars)
		})
	}
}
