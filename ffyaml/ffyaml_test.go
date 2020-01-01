package ffyaml_test

import (
	"testing"
	"time"

	"github.com/peterbourgon/ff"
	"github.com/peterbourgon/ff/fftest"
	"github.com/peterbourgon/ff/ffyaml"
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
			name: "no value",
			file: "testdata/no_value.yaml",
			want: fftest.Vars{WantParseErrorIs: ff.StringConversionError{}},
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
