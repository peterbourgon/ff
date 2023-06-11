package ff_test

import (
	"io"
	"testing"
	"time"

	"github.com/peterbourgon/ff/v3"
	"github.com/peterbourgon/ff/v3/fftest"
)

func TestJSONParser(t *testing.T) {
	t.Parallel()

	for _, testcase := range []struct {
		name string
		args []string
		file string
		want fftest.Vars
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
	} {
		t.Run(testcase.name, func(t *testing.T) {
			fs, vars := fftest.Pair()
			vars.ParseError = ff.Parse(fs, testcase.args,
				ff.WithConfigFile(testcase.file),
				ff.WithConfigFileParser(ff.JSONParser),
			)
			fftest.Compare(t, &testcase.want, vars)
		})
	}
}
