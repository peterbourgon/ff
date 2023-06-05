package ff_test

import (
	"flag"
	"io"
	"testing"
	"time"

	"github.com/peterbourgon/ff/v3"
	"github.com/peterbourgon/ff/v3/fftest"
)

func TestJSONParser(t *testing.T) {
	t.Parallel()

	for _, testcase := range []struct {
		vars func(*flag.FlagSet) *fftest.Vars
		name string
		args []string
		file string
		want fftest.Vars
		opts []ff.JSONOption
	}{
		{
			name: "empty input",
			args: []string{},
			file: "testdata/empty.json",
			want: fftest.Vars{},
		},
		{
			name: "basic KV pairs",
			args: []string{},
			file: "testdata/basic.json",
			want: fftest.Vars{S: "s", I: 10, B: true, D: 5 * time.Second},
		},
		{
			name: "value arrays",
			args: []string{},
			file: "testdata/value_arrays.json",
			want: fftest.Vars{S: "bb", I: 12, B: true, D: 5 * time.Second, X: []string{"a", "B", "👍"}},
		},
		{
			name: "bad JSON file",
			args: []string{},
			file: "testdata/bad.json",
			want: fftest.Vars{WantParseErrorIs: io.ErrUnexpectedEOF},
		},
		{
			vars: fftest.NestedDefaultVars("."),
			name: "nested objects",
			file: "testdata/nested.json",
			want: fftest.Vars{S: "a string", B: true, I: 123, F: 1.23, X: []string{"one", "two", "three"}},
		}, {
			vars: fftest.NestedDefaultVars("-"),
			name: "nested objects hyphen delimiter",
			file: "testdata/nested.json",
			want: fftest.Vars{S: "a string", B: true, I: 123, F: 1.23, X: []string{"one", "two", "three"}},
			opts: []ff.JSONOption{ff.WithJSONDelimiter("-")},
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
				ff.WithConfigFileParser(ff.NewJSONConfigFileParser(testcase.opts...).Parse),
			)
			fftest.Compare(t, &testcase.want, vars)
		})
	}
}
