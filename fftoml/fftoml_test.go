package fftoml_test

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/peterbourgon/ff"
	"github.com/peterbourgon/ff/fftoml"
)

func TestParser(t *testing.T) {
	type want struct {
		s   string
		i   int
		b   bool
		d   time.Duration
		f   float64
		err string
	}

	for _, testcase := range []struct {
		name string
		args []string
		file string
		want want
	}{
		{
			name: "empty input",
			args: []string{},
			file: ``,
			want: want{d: time.Second},
		},
		{
			name: "basic KV pairs",
			args: []string{},
			file: `
			s = "s"
			i = 10
			b = true
			d = "5s"
			f = 3.14e10
			`,
			want: want{"s", 10, true, 5 * time.Second, 3.14e10, ""},
		},
		{
			name: "bad TOML file",
			args: []string{},
			file: `{`,
			want: want{d: time.Second, err: "error parsing TOML config"},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			fs := flag.NewFlagSet("test", flag.ExitOnError)
			var (
				s = fs.String("s", "", "string")
				i = fs.Int("i", 0, "int")
				b = fs.Bool("b", false, "bool")
				d = fs.Duration("d", time.Second, "time.Duration")
				f = fs.Float64("f", 0.0, "float64")
			)

			var options []ff.Option
			{
				filename := filepath.Join(os.TempDir(), "TestParse"+fmt.Sprint(10000*rand.Intn(10000)))
				f, err := os.Create(filename)
				if err != nil {
					t.Fatal(err)
				}
				defer os.Remove(f.Name())
				f.Write([]byte(testcase.file))
				f.Close()

				options = append(options, ff.WithConfigFile(f.Name()), ff.WithConfigFileParser(fftoml.Parser))
			}

			err := ff.Parse(fs, testcase.args, options...)
			if testcase.want.err == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else {
				want, have := testcase.want.err, "<nil>"
				if err != nil {
					have = err.Error()
				}
				if !strings.Contains(have, want) {
					t.Errorf("missing expected error: want %q, have %q", want, have)
				}
			}

			if want, have := testcase.want.s, *s; want != have {
				t.Errorf("s: want %q, have %q", want, have)
			}
			if want, have := testcase.want.i, *i; want != have {
				t.Errorf("i: want %d, have %d", want, have)
			}
			if want, have := testcase.want.b, *b; want != have {
				t.Errorf("b: want %v, have %v", want, have)
			}
			if want, have := testcase.want.d, *d; want != have {
				t.Errorf("d: want %s, have %s", want, have)
			}
			if want, have := testcase.want.f, *f; want != have {
				t.Errorf("d: want %f, have %f", want, have)
			}
		})
	}
}
