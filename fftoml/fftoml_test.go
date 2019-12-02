package fftoml_test

import (
	"flag"
	"reflect"
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
			want: fftest.Vars{
				S: "s",
				I: 10,
				F: 3.14e10,
				B: true,
				D: 5 * time.Second,
				X: []string{"1", "a", "👍"},
			},
		},
		{
			name: "bad TOML file",
			file: "testdata/bad.toml",
			want: fftest.Vars{WantParseErrorString: "keys cannot contain { character"},
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

func TestParser_WithTables(t *testing.T) {
	var (
		tc struct {
			String  string
			Float   float64
			Strings fftest.StringSlice
		}
		fs = flag.NewFlagSet("fftest", flag.ContinueOnError)
	)

	fs.StringVar(&tc.String, "string-key", "", "string")
	fs.Float64Var(&tc.Float, "float-nested-key", 0, "float64")
	fs.Var(&tc.Strings, "strings-nested-key", "string slice")

	if err := ff.Parse(fs, []string{},
		ff.WithConfigFile("testdata/table.toml"),
		ff.WithConfigFileParser(fftoml.Parser)); err != nil {
		t.Fatal(err)
	}

	if tc.String != "a string" {
		t.Errorf(`expected string to be "a string", found %q`, tc.String)
	}

	if tc.Float != 1.23 {
		t.Errorf("expected float to be 1.23, found %v", tc.Float)
	}

	expected := fftest.StringSlice{"one", "two", "three"}
	if !reflect.DeepEqual(tc.Strings, expected) {
		t.Errorf("expected strings to be %q, found %q", expected, tc.Strings)
	}
}
