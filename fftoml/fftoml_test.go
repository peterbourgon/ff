package fftoml_test

import (
	"flag"
	"fmt"
	"reflect"
	"strings"
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
			fftest.Compare(t, &testcase.want, vars)
		})
	}
}

func TestParser_WithTables(t *testing.T) {
	t.Parallel()

	for _, delim := range []string{
		".",
		"-",
	} {
		t.Run(fmt.Sprintf("delim=%q", delim), func(t *testing.T) {
			var (
				skey = strings.Join([]string{"string", "key"}, delim)
				fkey = strings.Join([]string{"float", "nested", "key"}, delim)
				xkey = strings.Join([]string{"strings", "nested", "key"}, delim)

				sval string
				fval float64
				xval fftest.StringSlice
			)

			fs := flag.NewFlagSet("fftest", flag.ContinueOnError)
			{
				fs.StringVar(&sval, skey, "xxx", "string")
				fs.Float64Var(&fval, fkey, 999, "float64")
				fs.Var(&xval, xkey, "strings")
			}

			parseConfig := fftoml.New(fftoml.WithTableDelimiter(delim))

			if err := ff.Parse(fs, []string{},
				ff.WithConfigFile("testdata/table.toml"),
				ff.WithConfigFileParser(parseConfig.Parse),
			); err != nil {
				t.Fatal(err)
			}

			if want, have := "a string", sval; want != have {
				t.Errorf("string key: want %q, have %q", want, have)
			}

			if want, have := 1.23, fval; want != have {
				t.Errorf("float nested key: want %v, have %v", want, have)
			}

			if want, have := (fftest.StringSlice{"one", "two", "three"}), xval; !reflect.DeepEqual(want, have) {
				t.Errorf("strings nested key: want %v, have %v", want, have)
			}
		})
	}
}
