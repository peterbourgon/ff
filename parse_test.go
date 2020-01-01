package ff_test

import (
	"os"
	"testing"
	"time"

	"github.com/peterbourgon/ff"
	"github.com/peterbourgon/ff/fftest"
)

func TestParseBasics(t *testing.T) {
	t.Parallel()

	for _, testcase := range []struct {
		name string
		env  map[string]string
		file string
		args []string
		opts []ff.Option
		want fftest.Vars
	}{
		{
			name: "empty",
			args: []string{},
			want: fftest.Vars{},
		},
		{
			name: "args only",
			args: []string{"-s", "foo", "-i", "123", "-b", "-d", "24m"},
			want: fftest.Vars{S: "foo", I: 123, B: true, D: 24 * time.Minute},
		},
		{
			name: "file only",
			file: "testdata/1.conf",
			want: fftest.Vars{S: "bar", I: 99, B: true, D: time.Hour},
		},
		{
			name: "env only",
			env:  map[string]string{"TEST_PARSE_S": "baz", "TEST_PARSE_F": "0.99", "TEST_PARSE_D": "100s"},
			want: fftest.Vars{S: "baz", F: 0.99, D: 100 * time.Second},
		},
		{
			name: "file and args",
			file: "testdata/2.conf",
			args: []string{"-s", "foo", "-i", "1234"},
			want: fftest.Vars{S: "foo", I: 1234, D: 3 * time.Second},
		},
		{
			name: "env and args",
			env:  map[string]string{"TEST_PARSE_S": "should be overridden", "TEST_PARSE_B": "true"},
			args: []string{"-s", "explicit wins", "-i", "7"},
			want: fftest.Vars{S: "explicit wins", I: 7, B: true},
		},
		{
			name: "env and file",
			env:  map[string]string{"TEST_PARSE_S": "should be overridden", "TEST_PARSE_B": "true"},
			file: "testdata/3.conf",
			want: fftest.Vars{S: "bar", I: 99, B: true, D: 34 * time.Second},
		},
		{
			name: "env file args",
			env:  map[string]string{"TEST_PARSE_S": "from env", "TEST_PARSE_I": "300", "TEST_PARSE_F": "0.15", "TEST_PARSE_B": "true", "TEST_PARSE_D": "1h"},
			file: "testdata/4.conf",
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
			file: "testdata/5.conf",
			args: []string{"-s", "s.arg.1", "-s", "s.arg.2", "-x", "x.arg.1", "-x", "x.arg.2"},
			want: fftest.Vars{S: "s.arg.2", X: []string{"x.arg.1", "x.arg.2"}}, // highest prio wins and no others are called
		},
		{
			name: "PlainParser solo bool",
			file: "testdata/solo_bool.conf",
			want: fftest.Vars{S: "x", B: true},
		},
		{
			name: "PlainParser string with spaces",
			file: "testdata/spaces.conf",
			want: fftest.Vars{S: "i am the very model of a modern major general"},
		},
		{
			name: "default comma behavior",
			env:  map[string]string{"TEST_PARSE_S": "one,two,three", "TEST_PARSE_X": "one,two,three"},
			want: fftest.Vars{S: "three", X: []string{"one", "two", "three"}},
		},
		{
			name: "WithEnvVarIgnoreCommas",
			env:  map[string]string{"TEST_PARSE_S": "one,two,three", "TEST_PARSE_X": "one,two,three"},
			opts: []ff.Option{ff.WithEnvVarIgnoreCommas(true)},
			want: fftest.Vars{S: "one,two,three", X: []string{"one,two,three"}},
		},
		{
			name: "WithIgnoreUndefined env",
			env:  map[string]string{"TEST_PARSE_UNDEFINED": "one", "TEST_PARSE_S": "one"},
			opts: []ff.Option{ff.WithIgnoreUndefined(true)},
			want: fftest.Vars{S: "one"},
		},
		{
			name: "WithIgnoreUndefined file true",
			file: "testdata/undefined.conf",
			opts: []ff.Option{ff.WithIgnoreUndefined(true)},
			want: fftest.Vars{S: "one"},
		},
		{
			name: "WithIgnoreUndefined file false",
			file: "testdata/undefined.conf",
			opts: []ff.Option{ff.WithIgnoreUndefined(false)},
			want: fftest.Vars{WantParseErrorString: "config file flag"},
		},
		{
			name: "env var comma whitespace",
			env:  map[string]string{"TEST_PARSE_S": "one, two, three ", "TEST_PARSE_X": "one, two, three "},
			want: fftest.Vars{S: " three ", X: []string{"one", " two", " three "}},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			if testcase.file != "" {
				testcase.opts = append(testcase.opts, ff.WithConfigFile(testcase.file), ff.WithConfigFileParser(ff.PlainParser))
			}

			if len(testcase.env) > 0 {
				for k, v := range testcase.env {
					defer os.Setenv(k, os.Getenv(k))
					os.Setenv(k, v)
				}
				testcase.opts = append(testcase.opts, ff.WithEnvVarPrefix("TEST_PARSE"))
			}

			fs, vars := fftest.Pair()
			vars.ParseError = ff.Parse(fs, testcase.args, testcase.opts...)
			if err := fftest.Compare(&testcase.want, vars); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestParseIssue16(t *testing.T) {
	t.Parallel()

	for _, testcase := range []struct {
		name string
		data string
		want string
	}{
		{
			name: "hash in value",
			data: "s bar#baz",
			want: "bar#baz",
		},
		{
			name: "EOL comment with space",
			data: "s bar # baz",
			want: "bar",
		},
		{
			name: "EOL comment no space",
			data: "s bar #baz",
			want: "bar",
		},
		{
			name: "only comment with space",
			data: "# foo bar\n",
			want: "",
		},
		{
			name: "only comment no space",
			data: "#foo bar\n",
			want: "",
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			filename, cleanup := fftest.TempFile(t, testcase.data)
			defer cleanup()

			fs, vars := fftest.Pair()
			vars.ParseError = ff.Parse(fs, []string{},
				ff.WithConfigFile(filename),
				ff.WithConfigFileParser(ff.PlainParser),
			)

			want := fftest.Vars{S: testcase.want}
			if err := fftest.Compare(&want, vars); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestParseConfigFile(t *testing.T) {
	t.Parallel()

	for _, testcase := range []struct {
		name         string
		missing      bool
		allowMissing bool
		parseError   error
	}{
		{
			name: "has config file",
		},
		{
			name:       "config file missing",
			missing:    true,
			parseError: os.ErrNotExist,
		},
		{
			name:         "config file missing + allow missing",
			missing:      true,
			allowMissing: true,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			filename := "dummy"
			if !testcase.missing {
				var cleanup func()
				filename, cleanup = fftest.TempFile(t, "")
				defer cleanup()
			}

			options := []ff.Option{ff.WithConfigFile(filename), ff.WithConfigFileParser(ff.PlainParser)}
			if testcase.allowMissing {
				options = append(options, ff.WithAllowMissingConfigFile(true))
			}

			fs, vars := fftest.Pair()
			vars.ParseError = ff.Parse(fs, []string{}, options...)

			want := fftest.Vars{WantParseErrorIs: testcase.parseError}
			if err := fftest.Compare(&want, vars); err != nil {
				t.Fatal(err)
			}
		})
	}
}
