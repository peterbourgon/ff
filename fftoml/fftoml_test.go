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
			file: ``,
			want: fftest.Vars{},
		},
		{
			name: "basic KV pairs",
			file: `
			s = "s"
			i = 10
			f = 3.14e10
			b = true
			d = "5s"
			x = ["1", "a", "👍"]
			`,
			want: fftest.Vars{S: "s", I: 10, F: 3.14e10, B: true, D: 5 * time.Second, X: []string{"1", "a", "👍"}},
		},
		{
			name: "bad TOML file",
			file: `{`,
			want: fftest.Vars{WantParseErrorString: "bare keys cannot contain '{'"},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			filename, cleanup := fftest.TempFile(t, testcase.file)
			defer cleanup()

			fs, vars := fftest.Pair()
			vars.ParseError = ff.Parse(fs, []string{},
				ff.WithConfigFile(filename),
				ff.WithConfigFileParser(fftoml.Parser),
			)

			if err := fftest.Compare(&testcase.want, vars); err != nil {
				t.Fatal(err)
			}
		})
	}
}
