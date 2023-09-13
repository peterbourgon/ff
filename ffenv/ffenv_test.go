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

	testcases := fftest.TestCases{
		{
			ConfigFile: "testdata/empty.env",
			Want:       fftest.Vars{},
		},
		{
			ConfigFile: "testdata/basic.env",
			Want:       fftest.Vars{S: "bar", I: 99, B: true, D: time.Hour},
		},
		{
			ConfigFile: "testdata/prefix.env",
			Options:    []ff.Option{ff.WithEnvVarPrefix("MYPROG")},
			Want:       fftest.Vars{S: "bingo", I: 123},
		},
		{
			ConfigFile: "testdata/prefix-undef.env",
			Options:    []ff.Option{ff.WithEnvVarPrefix("MYPROG"), ff.WithConfigIgnoreUndefinedFlags()},
			Want:       fftest.Vars{S: "bango", I: 9},
		},
		{
			ConfigFile: "testdata/quotes.env",
			Want:       fftest.Vars{S: "", I: 32, X: []string{"1", "2 2", "3 3 3"}},
		},
		{
			ConfigFile: "testdata/no-value.env",
			Want:       fftest.Vars{WantParseErrorIs: ffenv.ErrInvalidLine},
		},
		{
			ConfigFile: "testdata/spaces.env",
			Want:       fftest.Vars{X: []string{"1", "2", "3", "4", "5", " 6", " 7 ", " 8 ", "9"}},
		},
		{
			ConfigFile: "testdata/newlines.env",
			Want:       fftest.Vars{S: "one\ntwo\nthree\n\n", X: []string{`A\nB\n\n`}},
		},
		{
			ConfigFile: "testdata/capitalization.env",
			Want:       fftest.Vars{S: "hello", I: 12345},
		},
		{
			ConfigFile: "testdata/comments.env",
			Want:       fftest.Vars{S: "abc # def"},
		},
	}

	for i := range testcases {
		if testcases[i].Name == "" {
			testcases[i].Name = filepath.Base(testcases[i].ConfigFile)
		}
	}

	testcases.Run(t)
}
