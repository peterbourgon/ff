package ff_test

import (
	"io"
	"testing"
	"time"

	"github.com/peterbourgon/ff"
	"github.com/peterbourgon/ff/fftest"
)

func TestJSONParser(t *testing.T) {
	for _, testcase := range []struct {
		name string
		args []string
		file string
		want fftest.Vars
	}{
		{
			name: "empty input",
			args: []string{},
			file: `{}`,
			want: fftest.Vars{D: time.Second},
		},
		{
			name: "basic KV pairs",
			args: []string{},
			file: `{"s": "s", "i": 10, "b": true, "d": "5s"}`,
			want: fftest.Vars{S: "s", I: 10, B: true, D: 5 * time.Second},
		},
		{
			name: "value arrays",
			args: []string{},
			file: `
				{
					"s": ["a", "bb"],
					"i": ["10", "11", "12"],
					"b": [false, true],
					"d": ["10m", "5s"],
					"x": ["a", "B", "👍"]
				}
			`,
			want: fftest.Vars{S: "bb", I: 12, B: true, D: 5 * time.Second, X: []string{"a", "B", "👍"}},
		},
		{
			name: "bad JSON file",
			args: []string{},
			file: `{`,
			want: fftest.Vars{D: 1 * time.Second, WantParseErrorIs: io.ErrUnexpectedEOF},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			filename, cleanup := fftest.TempFile(t, testcase.file)
			defer cleanup()

			fs, vars := fftest.Pair()
			vars.ParseError = ff.Parse(fs, testcase.args,
				ff.WithConfigFile(filename),
				ff.WithConfigFileParser(ff.JSONParser),
			)

			if err := fftest.Compare(&testcase.want, vars); err != nil {
				t.Fatal(err)
			}
		})
	}
}
