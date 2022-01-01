package ff_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/peterbourgon/ff/v3"
	"github.com/peterbourgon/ff/v3/fftest"
)

func TestEnvFileParser(t *testing.T) {
	t.Parallel()

	for _, testcase := range []struct {
		file string
		opts []ff.Option
		want fftest.Vars
	}{
		{
			file: "testdata/empty.env",
			want: fftest.Vars{},
		},
		{
			file: "testdata/basic.env",
			want: fftest.Vars{S: "bar", I: 99, B: true, D: time.Hour},
		},
		{
			file: "testdata/prefix.env",
			opts: []ff.Option{ff.WithEnvVarPrefix("MYPROG")},
			want: fftest.Vars{S: "bingo", I: 123},
		},
		{
			file: "testdata/prefix-undef.env",
			opts: []ff.Option{ff.WithEnvVarPrefix("MYPROG"), ff.WithIgnoreUndefined(true)},
			want: fftest.Vars{S: "bango", I: 9},
		},
		{
			file: "testdata/quotes.env",
			want: fftest.Vars{S: "", I: 32, X: []string{"1", "2 2", "3 3 3"}},
		},
		{
			file: "testdata/no-value.env",
			want: fftest.Vars{WantParseErrorString: "invalid line: D="},
		},
		{
			file: "testdata/spaces.env",
			want: fftest.Vars{X: []string{"1", "2", "3", "4", "5", " 6", " 7 ", " 8 "}},
		},
		{
			file: "testdata/newlines.env",
			want: fftest.Vars{S: "one\ntwo\nthree\n\n", X: []string{`A\nB\n\n`}},
		},
		{
			file: "testdata/capitalization.env",
			want: fftest.Vars{S: "hello", I: 12345},
		},
	} {
		t.Run(filepath.Base(testcase.file), func(t *testing.T) {
			testcase.opts = append(testcase.opts, ff.WithConfigFile(testcase.file), ff.WithConfigFileParser(ff.EnvParser))
			fs, vars := fftest.Pair()
			vars.ParseError = ff.Parse(fs, []string{}, testcase.opts...)
			fftest.Compare(t, &testcase.want, vars)
		})
	}
}
