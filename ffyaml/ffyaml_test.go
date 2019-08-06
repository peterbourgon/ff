package ffyaml_test

import (
	"testing"
	"time"

	"github.com/peterbourgon/ff"
	"github.com/peterbourgon/ff/fftest"
	"github.com/peterbourgon/ff/ffyaml"
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
			file: "s: hello\ni: 10\nb: true\nd: 5s\nf: 3.14",
			want: fftest.Vars{S: "hello", I: 10, B: true, D: 5 * time.Second, F: 3.14},
		},
		{
			name: "invalid prefix",
			file: "\ti: 123\ns: foo\n",
			want: fftest.Vars{WantParseErrorString: "found character that cannot start any token"},
		},
		{
			name: "no value",
			file: "i: 123\ns:\n",
			want: fftest.Vars{I: 123, WantParseErrorIs: ff.StringConversionError{}},
		},
		{
			name: "no file",
			file: ``,
			want: fftest.Vars{},
		},
		{
			name: "basic arrays",
			file: "s: ['a', 'b', 'c']\n\nx: ['a', 'b', 'c']",
			want: fftest.Vars{S: "c", X: []string{"a", "b", "c"}},
		},
		{
			name: "multiline arrays",
			file: "s:\n  - a\n  - b\n  - c\nx:\n  - d\n  - e\n  - f\n",
			want: fftest.Vars{S: "c", X: []string{"d", "e", "f"}},
		},
		{
			name: "line break arrays",
			file: "x: [\"first string\", \"second\n string\", \"third\"]\n",
			want: fftest.Vars{X: []string{"first string", "second string", "third"}},
		},
		{
			name: "unquoted strings in arrays",
			file: "x: [one, two, three]",
			want: fftest.Vars{X: []string{"one", "two", "three"}},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			filename, cleanup := fftest.TempFile(t, testcase.file)
			defer cleanup()

			fs, vars := fftest.Pair()
			vars.ParseError = ff.Parse(fs, []string{},
				ff.WithConfigFile(filename),
				ff.WithConfigFileParser(ffyaml.Parser),
			)

			if err := fftest.Compare(&testcase.want, vars); err != nil {
				t.Fatal(err)
			}
		})
	}
}
