package ff

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestParsePriority(t *testing.T) {
	type want struct {
		s  string
		i  int
		b  bool
		d  time.Duration
		ss strs
	}

	for _, testcase := range []struct {
		name string
		args []string
		file string
		env  map[string]string
		want want
	}{
		{
			name: "args only",
			args: []string{"-s", "foo", "-i", "123", "-b", "-d", "24m", "-ss", "foo,bar"},
			want: want{"foo", 123, true, 24 * time.Minute, strs{"foo", "bar"}},
		},
		{
			name: "file only",
			file: "s bar\ni 99\nb true\nd 1h\nss 1,2,3",
			want: want{"bar", 99, true, time.Hour, strs{"1", "2", "3"}},
		},
		{
			name: "env only",
			env:  map[string]string{"TEST_PARSE_S": "baz", "TEST_PARSE_D": "100s", "TEST_PARSE_SS": "a,b,c"},
			want: want{"baz", 0, false, 100 * time.Second, strs{"a", "b","c"}},
		},
		{
			name: "args and file",
			args: []string{"-s", "foo", "-i", "1234", "-ss", "foo,bar"},
			file: "\ns should be overridden\n\nd 3s\n",
			want: want{"foo", 1234, false, 3 * time.Second, strs{"foo", "bar"}},
		},
		{
			name: "args and env",
			args: []string{"-s", "explicit wins", "-i", "7"},
			env:  map[string]string{"TEST_PARSE_S": "should be overridden", "TEST_PARSE_B": "true", "TEST_PARSE_SS": "foo,bar"},
			want: want{"explicit wins", 7, true, time.Second, strs{"foo", "bar"}},
		},
		{
			name: "file and env",
			file: "s bar\ni 99\n\nd 34s\n\n # comment line\nss foo,bar\n",
			env:  map[string]string{"TEST_PARSE_S": "should be overridden", "TEST_PARSE_B": "true"},
			want: want{"bar", 99, true, 34 * time.Second, strs{"foo", "bar"}},
		},
		{
			name: "args file env",
			args: []string{"-s", "from arg", "-i", "100", "-ss", "foo,bar"},
			file: "s from file\ni 200 # comment\n\nd 1m\n\n\n",
			env:  map[string]string{"TEST_PARSE_S": "from env", "TEST_PARSE_I": "300", "TEST_PARSE_B": "true", "TEST_PARSE_D": "1h"},
			want: want{"from arg", 100, true, time.Minute, strs{"foo", "bar"}},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			fs := flag.NewFlagSet("test", flag.ExitOnError)
			var (
				s  = fs.String("s", "", "string")
				i  = fs.Int("i", 0, "int")
				b  = fs.Bool("b", false, "bool")
				d  = fs.Duration("d", time.Second, "time.Duration")
				ss strs
			)
			fs.Var(&ss, "ss", "comma-separated strings")

			var options []Option

			if testcase.file != "" {
				filename := filepath.Join(os.TempDir(), "TestParse"+fmt.Sprint(10000*rand.Intn(10000)))
				f, err := os.Create(filename)
				if err != nil {
					t.Fatal(err)
				}
				defer os.Remove(f.Name())
				f.Write([]byte(testcase.file))
				f.Close()

				options = append(options, WithConfigFile(f.Name()), WithConfigFileParser(PlainParser))
			}

			if len(testcase.env) > 0 {
				for k, v := range testcase.env {
					os.Setenv(k, v)
					defer os.Setenv(k, "")
				}

				options = append(options, WithEnvVarPrefix("TEST_PARSE"))
			}

			if err := Parse(fs, testcase.args, options...); err != nil {
				t.Fatal(err)
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
			if want, have := testcase.want.ss, ss; want.String() != have.String() {
				t.Errorf("ss: want %s, have %s", want, have)
			}
		})
	}
}

// strs is a slice of strings that implements the flag.Value interface so that it can be set with the
// parsing functions for flags, files and environment variables.
type strs []string

// String is the method to format the flag's value, part of the flag.Value interface.
// The String method's output will be used in diagnostics.
func (ii *strs) String() string {
	return fmt.Sprintf("%#v", ii)
}

// Set is the method to set the flag value, part of the flag.Value interface.
// Set's argument is a string to be parsed to set the flag.
func (ii *strs) Set(value string) error {
	if len(strings.TrimSpace(value)) == 0 {
		return nil
	}

	*ii = strings.Split(value, ",")

	return nil
}
