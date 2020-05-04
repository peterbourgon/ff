package fftoml_test

import (
	"flag"
	"reflect"
	"testing"
	"time"

	"github.com/peterbourgon/ff/v3"
	"github.com/peterbourgon/ff/v3/fftest"
	"github.com/peterbourgon/ff/v3/fftoml"
)

func TestParser(t *testing.T) {
	t.Parallel()

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

			pfs, pvars := fftest.PairPflag()
			pvars.ParseError = ff.Parse(ff.FromPflag(pfs), []string{},
				ff.WithConfigFile(testcase.file),
				ff.WithConfigFileParser(fftoml.Parser),
			)
			if err := fftest.Compare(&testcase.want, pvars); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestParser_WithTables(t *testing.T) {
	t.Parallel()

	type fields struct {
		String  string
		Float   float64
		Strings fftest.StringSlice
	}

	expected := fields{
		String:  "a string",
		Float:   1.23,
		Strings: fftest.StringSlice{"one", "two", "three"},
	}

	for _, testcase := range []struct {
		name string
		opts []fftoml.Option
		// expectations
		stringKey  string
		floatKey   string
		stringsKey string
	}{
		{
			name:       "defaults",
			stringKey:  "string.key",
			floatKey:   "float.nested.key",
			stringsKey: "strings.nested.key",
		},
		{
			name:       "defaults",
			opts:       []fftoml.Option{fftoml.WithTableDelimiter("-")},
			stringKey:  "string-key",
			floatKey:   "float-nested-key",
			stringsKey: "strings-nested-key",
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			var (
				found fields
				fs    = flag.NewFlagSet("fftest", flag.ContinueOnError)
			)

			fs.StringVar(&found.String, testcase.stringKey, "", "string")
			fs.Float64Var(&found.Float, testcase.floatKey, 0, "float64")
			fs.Var(&found.Strings, testcase.stringsKey, "string slice")

			if err := ff.Parse(fs, []string{},
				ff.WithConfigFile("testdata/table.toml"),
				ff.WithConfigFileParser(fftoml.New(testcase.opts...).Parse),
			); err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(expected, found) {
				t.Errorf(`expected %v, to be %v`, found, expected)
			}
		})
	}

}
