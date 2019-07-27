package ff_test

import (
	"os"
	"testing"
	"time"

	"github.com/peterbourgon/ff"
	"github.com/peterbourgon/ff/fftest"
)

func TestParseBasics(t *testing.T) {
	for _, testcase := range []struct {
		name string
		args []string
		file string
		env  map[string]string
		want fftest.Vars
	}{
		{
			name: "args only",
			args: []string{"-s", "foo", "-i", "123", "-b", "-d", "24m"},
			want: fftest.Vars{S: "foo", I: 123, B: true, D: 24 * time.Minute},
		},
		{
			name: "file only",
			file: "s bar\ni 99\nb true\nd 1h",
			want: fftest.Vars{S: "bar", I: 99, B: true, D: time.Hour},
		},
		{
			name: "env only",
			env:  map[string]string{"TEST_PARSE_S": "baz", "TEST_PARSE_F": "0.99", "TEST_PARSE_D": "100s"},
			want: fftest.Vars{S: "baz", F: 0.99, D: 100 * time.Second},
		},
		{
			name: "args and file",
			args: []string{"-s", "foo", "-i", "1234"},
			file: "\ns should be overridden\n\nd 3s\n",
			want: fftest.Vars{S: "foo", I: 1234, D: 3 * time.Second},
		},
		{
			name: "args and env",
			args: []string{"-s", "explicit wins", "-i", "7"},
			env:  map[string]string{"TEST_PARSE_S": "should be overridden", "TEST_PARSE_B": "true"},
			want: fftest.Vars{S: "explicit wins", I: 7, B: true, D: time.Second},
		},
		{
			name: "file and env",
			file: "s bar\ni 99\n\nd 34s\n\n # comment line\n",
			env:  map[string]string{"TEST_PARSE_S": "should be overridden", "TEST_PARSE_B": "true"},
			want: fftest.Vars{S: "bar", I: 99, B: true, D: 34 * time.Second},
		},
		{
			name: "args file env",
			args: []string{"-s", "from arg", "-i", "100"},
			file: "s from file\ni 200 # comment\n\nd 1m\nf 2.3\n\n",
			env:  map[string]string{"TEST_PARSE_S": "from env", "TEST_PARSE_I": "300", "TEST_PARSE_F": "0.15", "TEST_PARSE_B": "true", "TEST_PARSE_D": "1h"},
			want: fftest.Vars{S: "from arg", I: 100, F: 2.3, B: true, D: time.Minute},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			var options []ff.Option

			if testcase.file != "" {
				filename, cleanup := fftest.CreateTempFile(t, testcase.file)
				defer cleanup()
				options = append(options, ff.WithConfigFile(filename), ff.WithConfigFileParser(ff.PlainParser))
			}

			if len(testcase.env) > 0 {
				for k, v := range testcase.env {
					os.Setenv(k, v)
					defer os.Setenv(k, "")
				}
				options = append(options, ff.WithEnvVarPrefix("TEST_PARSE"))
			}

			fs, vars := fftest.NewPair()
			vars.ParseError = ff.Parse(fs, testcase.args, options...)
			if err := fftest.Compare(&testcase.want, vars); err != nil {
				t.Fatal(err)
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
			file: "s bar#baz",
			want: "bar#baz",
		},
		{
			name: "EOL comment with space",
			file: "s bar # baz",
			want: "bar",
		},
		{
			name: "EOL comment no space",
			file: "s bar #baz",
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
			filename, cleanup := fftest.CreateTempFile(t, testcase.file)
			defer cleanup()

			fs, vars := fftest.NewPair()
			vars.ParseError = ff.Parse(fs, []string{},
				ff.WithConfigFile(filename),
				ff.WithConfigFileParser(ff.PlainParser),
			)
			want := fftest.Vars{S: testcase.want, D: time.Second}
			if err := fftest.Compare(&want, vars); err != nil {
				t.Fatal(err)
			}
		})
	}
}
