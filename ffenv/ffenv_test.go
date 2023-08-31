package ffenv_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffenv"
	"github.com/peterbourgon/ff/v4/fftest"
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
			opts: []ff.Option{ff.WithEnvVarPrefix("MYPROG"), ff.WithConfigIgnoreUndefinedFlags()},
			want: fftest.Vars{S: "bango", I: 9},
		},
		{
			file: "testdata/quotes.env",
			want: fftest.Vars{S: "", I: 32, X: []string{"1", "2 2", "3 3 3"}},
		},
		{
			file: "testdata/no-value.env",
			want: fftest.Vars{WantParseErrorIs: ffenv.ErrInvalidLine},
		},
		{
			file: "testdata/spaces.env",
			want: fftest.Vars{X: []string{"1", "2", "3", "4", "5", " 6", " 7 ", " 8 ", "9"}},
		},
		{
			file: "testdata/newlines.env",
			want: fftest.Vars{S: "one\ntwo\nthree\n\n", X: []string{`A\nB\n\n`}},
		},
		{
			file: "testdata/capitalization.env",
			want: fftest.Vars{S: "hello", I: 12345},
		},
		{
			file: "testdata/comments.env",
			want: fftest.Vars{S: "abc # def"},
		},
	} {
		t.Run(filepath.Base(testcase.file), func(t *testing.T) {
			testcase.opts = append(testcase.opts,
				ff.WithConfigFile(testcase.file),
				ff.WithConfigFileParser(ffenv.Parse),
			)
			for _, constr := range fftest.DefaultConstructors {
				t.Run(constr.Name, func(t *testing.T) {
					fs, vars := constr.Make(fftest.Vars{})
					vars.ParseError = ff.Parse(fs, []string{}, testcase.opts...)
					fftest.Compare(t, &testcase.want, vars)
				})
			}
		})
	}
}
