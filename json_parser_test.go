package ff_test

import (
	"flag"
	"io"
	"reflect"
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

func TestParser_WithNested(t *testing.T) {
	t.Parallel()

	type fields struct {
		String  string
		Bool    bool
		Float   float64
		Strings fftest.StringSlice
	}

	expected := fields{
		String:  "a string",
		Bool:    true,
		Float:   1.23,
		Strings: fftest.StringSlice{"one", "two", "three"},
	}

	for _, testcase := range []struct {
		name string
		opts []ff.JSONOption
		// expectations
		stringKey  string
		boolKey    string
		floatKey   string
		stringsKey string
	}{
		{
			name:       "defaults",
			stringKey:  "string.key",
			boolKey:    "string.bool",
			floatKey:   "float.nested.key",
			stringsKey: "strings.nested.key",
		},
		{
			name:       "delimiter",
			opts:       []ff.JSONOption{ff.WithObjectDelimiter("-")},
			stringKey:  "string-key",
			boolKey:    "string-bool",
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
			fs.BoolVar(&found.Bool, testcase.boolKey, false, "bool")
			fs.Float64Var(&found.Float, testcase.floatKey, 0, "float64")
			fs.Var(&found.Strings, testcase.stringsKey, "string slice")

			if err := ff.Parse(fs, []string{},
				ff.WithConfigFile("testdata/nested.json"),
				ff.WithConfigFileParser(ff.NewJSONParser(testcase.opts...).Parse),
			); err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(expected, found) {
				t.Errorf(`expected %v, to be %v`, found, expected)
			}
		})
	}
}
