package fftoml_test

import (
	"testing"
	"time"

	"github.com/peterbourgon/ff"
	"github.com/peterbourgon/ff/fftest"
	"github.com/peterbourgon/ff/fftoml"
)

func TestParser(t *testing.T) {
	for _, testcase := range []struct {
		name string
		file string
		want fftest.Vars
	}{
		{
			name: "empty input",
			file: "testdata/empty.toml",
			want: fftest.Vars{},
		},
		{
			name: "basic KV pairs",
			file: "testdata/basic.toml",
			want: fftest.Vars{S: "s", I: 10, F: 3.14e10, B: true, D: 5 * time.Second, X: []string{"1", "a", "👍"}},
		},
		{
			name: "bad TOML file",
			file: "testdata/bad.toml",
			want: fftest.Vars{WantParseErrorString: "bare keys cannot contain '{'"},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			fs, vars := fftest.Pair()
			vars.ParseError = ff.Parse(fs, []string{},
				ff.WithConfigFile(testcase.file),
				ff.WithConfigFileParser(fftoml.Parser),
			)
			if err := fftest.Compare(&testcase.want, vars); err != nil {
				t.Fatal(err)
			}
		})
	}
}
