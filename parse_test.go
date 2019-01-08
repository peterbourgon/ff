package ff

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParsePriority(t *testing.T) {
	type want struct {
		s string
		i int
		b bool
		d time.Duration
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
			args: []string{"-s", "foo", "-i", "123", "-b", "-d", "24m"},
			want: want{"foo", 123, true, 24 * time.Minute},
		},
		{
			name: "file only",
			file: "s bar\ni 99\nb true\nd 1h",
			want: want{"bar", 99, true, time.Hour},
		},
		{
			name: "env only",
			env:  map[string]string{"TEST_PARSE_S": "baz", "TEST_PARSE_D": "100s"},
			want: want{"baz", 0, false, 100 * time.Second},
		},
		{
			name: "args and file",
			args: []string{"-s", "foo", "-i", "1234"},
			file: "\ns should be overridden\n\nd 3s\n",
			want: want{"foo", 1234, false, 3 * time.Second},
		},
		{
			name: "args and env",
			args: []string{"-s", "explicit wins", "-i", "7"},
			env:  map[string]string{"TEST_PARSE_S": "should be overridden", "TEST_PARSE_B": "true"},
			want: want{"explicit wins", 7, true, time.Second},
		},
		{
			name: "file and env",
			file: "s bar\ni 99\n\nd 34s\n\n # comment line\n",
			env:  map[string]string{"TEST_PARSE_S": "should be overridden", "TEST_PARSE_B": "true"},
			want: want{"bar", 99, true, 34 * time.Second},
		},
		{
			name: "args file env",
			args: []string{"-s", "from arg", "-i", "100"},
			file: "s from file\ni 200 # comment\n\nd 1m\n\n\n",
			env:  map[string]string{"TEST_PARSE_S": "from env", "TEST_PARSE_I": "300", "TEST_PARSE_B": "true", "TEST_PARSE_D": "1h"},
			want: want{"from arg", 100, true, time.Minute},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			fs := flag.NewFlagSet("test", flag.ExitOnError)
			var (
				s = fs.String("s", "", "string")
				i = fs.Int("i", 0, "int")
				b = fs.Bool("b", false, "bool")
				d = fs.Duration("d", time.Second, "time.Duration")
			)

			var options []Option

			if testcase.file != "" {
				filename := filepath.Join(os.TempDir(), "TestParsePriority"+fmt.Sprint(10000*rand.Intn(10000)))
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
		})
	}
}

func TestParseIssue16(t *testing.T) {
	for _, testcase := range []struct {
		name string
		file string
		want string
	}{
		{
			name: "hash in value",
			file: "foo bar#baz",
			want: "bar#baz",
		},
		{
			name: "EOL comment with space",
			file: "foo bar # baz",
			want: "bar",
		},
		{
			name: "EOL comment no space",
			file: "foo bar #baz",
			want: "bar",
		},
		{
			name: "only comment with space",
			file: "# foo bar\n",
			want: "",
		},
		{
			name: "only comment no space",
			file: "#foo bar\n",
			want: "",
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			fs := flag.NewFlagSet("test", flag.ExitOnError)
			foo := fs.String("foo", "", "the value of foo")

			filename := filepath.Join(os.TempDir(), "TestParseIssue16"+fmt.Sprint(10000*rand.Intn(10000)))
			f, err := os.Create(filename)
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(f.Name())
			f.Write([]byte(testcase.file))
			f.Close()

			if err := Parse(fs, []string{}, WithConfigFile(filename), WithConfigFileParser(PlainParser)); err != nil {
				t.Fatal(err)
			}

			if want, have := testcase.want, *foo; want != have {
				t.Errorf("want %q, have %q", want, have)
			}
		})
	}
}
