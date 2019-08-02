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
		env  map[string]string
		file string
		args []string
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
			name: "file and args",
			file: "\ns should be overridden\n\nd 3s\n",
			args: []string{"-s", "foo", "-i", "1234"},
			want: fftest.Vars{S: "foo", I: 1234, D: 3 * time.Second},
		},
		{
			name: "env and args",
			env:  map[string]string{"TEST_PARSE_S": "should be overridden", "TEST_PARSE_B": "true"},
			args: []string{"-s", "explicit wins", "-i", "7"},
			want: fftest.Vars{S: "explicit wins", I: 7, B: true, D: time.Second},
		},
		{
			name: "env and file",
			env:  map[string]string{"TEST_PARSE_S": "should be overridden", "TEST_PARSE_B": "true"},
			file: "s bar\ni 99\n\nd 34s\n\n # comment line\n",
			want: fftest.Vars{S: "bar", I: 99, B: true, D: 34 * time.Second},
		},
		{
			name: "env file args",
			env:  map[string]string{"TEST_PARSE_S": "from env", "TEST_PARSE_I": "300", "TEST_PARSE_F": "0.15", "TEST_PARSE_B": "true", "TEST_PARSE_D": "1h"},
			file: "s from file\ni 200 # comment\n\nd 1m\nf 2.3\n\n",
			args: []string{"-s", "from arg", "-i", "100"},
			want: fftest.Vars{S: "from arg", I: 100, F: 2.3, B: true, D: time.Minute},
		},
		{
			name: "repeated args",
			args: []string{"-s", "foo", "-s", "bar", "-d", "1m", "-d", "1h", "-x", "1", "-x", "2", "-x", "3"},
			want: fftest.Vars{S: "bar", D: time.Hour, X: []string{"1", "2", "3"}},
		},
		{
			name: "priority repeats",
			env:  map[string]string{"TEST_PARSE_S": "s.env", "TEST_PARSE_X": "x.env.1"},
			file: "s s.file.1\ns s.file.2\n\nx x.file.1\nx x.file.2",
			args: []string{"-s", "s.arg.1", "-s", "s.arg.2", "-x", "x.arg.1", "-x", "x.arg.2"},
			want: fftest.Vars{S: "s.arg.2", D: time.Second, X: []string{"x.arg.1", "x.arg.2"}}, // highest prio wins and no others are called
		},
		{
			name: "PlainParser solo bool",
			file: "b\ns x\n",
			want: fftest.Vars{S: "x", D: time.Second, B: true},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			var options []ff.Option

			if testcase.file != "" {
				filename, cleanup := fftest.TempFile(t, testcase.file)
				defer cleanup()
				options = append(options, ff.WithConfigFile(filename), ff.WithConfigFileParser(ff.PlainParser))
			}

			if len(testcase.env) > 0 {
				for k, v := range testcase.env {
					defer os.Setenv(k, os.Getenv(k))
					os.Setenv(k, v)
				}
				options = append(options, ff.WithEnvVarPrefix("TEST_PARSE"))
			}

			fs, vars := fftest.Pair()
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
			filename, cleanup := fftest.TempFile(t, testcase.file)
			defer cleanup()

			fs, vars := fftest.Pair()
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
