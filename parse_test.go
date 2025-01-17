package ff_test

import (
	"embed"
	"flag"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/fftest"
)

//go:embed testdata/*.conf
var testdataConfigFS embed.FS

func TestParse(t *testing.T) {
	t.Parallel()

	testcases := fftest.TestCases{
		{
			Name: "empty",
			Want: fftest.Vars{},
		},
		{
			Name: "args",
			Args: []string{"-s", "foo", "-i", "123", "-b", "-d", "24m"},
			Want: fftest.Vars{S: "foo", I: 123, B: true, D: 24 * time.Minute},
		},

		{
			Name:       "file only",
			ConfigFile: "testdata/1.conf",
			Want:       fftest.Vars{S: "bar", I: 99, B: true, D: time.Hour},
		},
		{
			Name:        "env only",
			Environment: map[string]string{"TEST_PARSE_S": "baz", "TEST_PARSE_F": "0.99", "TEST_PARSE_D": "100s"},
			Options:     []ff.Option{ff.WithEnvVarPrefix("TEST_PARSE")},
			Want:        fftest.Vars{S: "baz", F: 0.99, D: 100 * time.Second},
		},
		{
			Name:       "file args",
			ConfigFile: "testdata/2.conf",
			Args:       []string{"-s", "foo", "-i", "1234"},
			Want:       fftest.Vars{S: "foo", I: 1234, D: 3 * time.Second},
		},
		{
			Name:        "env args",
			Environment: map[string]string{"TEST_PARSE_S": "should be overridden", "TEST_PARSE_B": "true"},
			Args:        []string{"-s", "explicit wins", "-i", "7"},
			Options:     []ff.Option{ff.WithEnvVarPrefix("TEST_PARSE")},
			Want:        fftest.Vars{S: "explicit wins", I: 7, B: true},
		},
		{
			Name:        "file env",
			ConfigFile:  "testdata/3.conf",
			Environment: map[string]string{"TEST_PARSE_S": "env takes priority", "TEST_PARSE_B": "true"},
			Options:     []ff.Option{ff.WithEnvVarPrefix("TEST_PARSE")},
			Want:        fftest.Vars{S: "env takes priority", I: 99, B: true, D: 34 * time.Second},
		},
		{
			Name:        "file env args",
			ConfigFile:  "testdata/4.conf",
			Environment: map[string]string{"TEST_PARSE_S": "from env", "TEST_PARSE_I": "300", "TEST_PARSE_F": "0.15", "TEST_PARSE_B": "true"},
			Args:        []string{"-s", "from arg", "-i", "100"},
			Options:     []ff.Option{ff.WithEnvVarPrefix("TEST_PARSE")},
			Want:        fftest.Vars{S: "from arg", I: 100, F: 0.15, B: true, D: time.Minute},
		},
		{
			Name: "repeated args",
			Args: []string{"-s", "foo", "-s", "bar", "-d", "1m", "-d", "1h", "-x", "1", "-x", "2", "-x", "3"},
			Want: fftest.Vars{S: "bar", D: time.Hour, X: []string{"1", "2", "3"}},
		},
		{
			Name:       "file repeats",
			ConfigFile: "testdata/5.conf",
			Want:       fftest.Vars{S: "s.file.2", X: []string{"x.file.1", "x.file.2"}},
		},
		{
			Name:        "priority repeats",
			ConfigFile:  "testdata/5.conf",
			Environment: map[string]string{"TEST_PARSE_S": "s.env", "TEST_PARSE_X": "x.env.1"},
			Args:        []string{"-s", "s.arg.1", "-s", "s.arg.2", "-x", "x.arg.1", "-x", "x.arg.2"},
			Options:     []ff.Option{ff.WithEnvVarPrefix("TEST_PARSE")},
			Want:        fftest.Vars{S: "s.arg.2", X: []string{"x.arg.1", "x.arg.2"}}, // highest prio wins and no others are called
		},
		{
			Name:        "WithEnvVars",
			Environment: map[string]string{"S": "xxx", "F": "9.87"},
			Options:     []ff.Option{ff.WithEnvVars()},
			Want:        fftest.Vars{S: "xxx", F: 9.87},
		},
		{
			Name:        "WithEnvVars prefix",
			Environment: map[string]string{"TEST_PARSE_S": "foo", "S": "bar"},
			Options:     []ff.Option{ff.WithEnvVarPrefix("TEST_PARSE")},
			Want:        fftest.Vars{S: "foo"},
		},
		{
			Name:        "WithEnvVars no prefix",
			Environment: map[string]string{"TEST_PARSE_S": "foo", "S": "bar"},
			Options:     []ff.Option{ff.WithEnvVars()},
			Want:        fftest.Vars{S: "bar"},
		},
		{
			Name:        "WithEnvVarSplit",
			Environment: map[string]string{"TEST_PARSE_S": "one,two,three", "TEST_PARSE_X": "one,two,three"},
			Options:     []ff.Option{ff.WithEnvVarPrefix("TEST_PARSE"), ff.WithEnvVarSplit(",")},
			Want:        fftest.Vars{S: "three", X: []string{"one", "two", "three"}},
		},
		{
			Name:        "env default comma behavior",
			Environment: map[string]string{"TEST_PARSE_S": "one,two,three", "TEST_PARSE_X": "one,two,three"},
			Options:     []ff.Option{ff.WithEnvVarPrefix("TEST_PARSE")},
			Want:        fftest.Vars{S: "one,two,three", X: []string{"one,two,three"}},
		},
		{
			Name:        "env var split comma whitespace",
			Environment: map[string]string{"TEST_PARSE_S": "one, two, three ", "TEST_PARSE_X": "one, two, three "},
			Options:     []ff.Option{ff.WithEnvVarPrefix("TEST_PARSE"), ff.WithEnvVarSplit(",")},
			Want:        fftest.Vars{S: " three ", X: []string{"one", " two", " three "}},
		},
		{
			Name:        "env var split escaping",
			Environment: map[string]string{"TEST_PARSE_S": `a\,b`, "TEST_PARSE_X": `one,two\,three`},
			Options:     []ff.Option{ff.WithEnvVarPrefix("TEST_PARSE"), ff.WithEnvVarSplit(",")},
			Want:        fftest.Vars{S: `a,b`, X: []string{`one`, `two,three`}},
		},
		{
			Name:        "env var split escaping multichar",
			Environment: map[string]string{"TEST_PARSE_S": `a\xxb`, "TEST_PARSE_X": `onexxtwo\xxthree`},
			Options:     []ff.Option{ff.WithEnvVarPrefix("TEST_PARSE"), ff.WithEnvVarSplit("xx")},
			Want:        fftest.Vars{S: `axxb`, X: []string{`one`, `twoxxthree`}},
		},
	}

	testcases.Run(t)
}

func TestParse_FlagSet(t *testing.T) {
	t.Parallel()

	testcases := fftest.TestCases{
		{
			Name:         "long args",
			Constructors: []fftest.Constructor{fftest.CoreConstructor},
			Args:         []string{"--str=foo", "--int", "123", "--bflag", "-d", "13m"},
			Want:         fftest.Vars{S: "foo", I: 123, B: true, D: 13 * time.Minute},
		},
		{
			Name:         "-b only",
			Constructors: []fftest.Constructor{fftest.CoreConstructor},
			Args:         []string{`-b`},
			Want:         fftest.Vars{B: true},
		},
		{
			Name:         "--str abc",
			Constructors: []fftest.Constructor{fftest.CoreConstructor},
			Args:         []string{`--str`, `abc`},
			Want:         fftest.Vars{S: "abc"},
		},
		{
			Name:         "-s xxx",
			Constructors: []fftest.Constructor{fftest.CoreConstructor},
			Args:         []string{`-s`, `xxx`},
			Want:         fftest.Vars{S: "xxx"},
		},
		{
			Name:         "-s=xxx",
			Constructors: []fftest.Constructor{fftest.CoreConstructor},
			Args:         []string{`-s=xxx`},
			Want:         fftest.Vars{S: "=xxx"},
		},
		{
			Name:         "-str=xxx",
			Constructors: []fftest.Constructor{fftest.CoreConstructor},
			Args:         []string{`-str=xxx`},
			Want:         fftest.Vars{S: `tr=xxx`},
		},
		{
			Name:         "-s -b",
			Constructors: []fftest.Constructor{fftest.CoreConstructor},
			Args:         []string{`-s`, `-b`},
			Want:         fftest.Vars{S: "-b"},
		},
		{
			Name:         "-a -b -c",
			Constructors: []fftest.Constructor{fftest.CoreConstructor},
			Args:         []string{`-a`, `-b`, `-c`},
			Want:         fftest.Vars{A: true, B: true, C: true},
		},
		{
			Name:         "-ab -c",
			Constructors: []fftest.Constructor{fftest.CoreConstructor},
			Args:         []string{`-ab`, `-c`},
			Want:         fftest.Vars{A: true, B: true, C: true},
		},
		{
			Name:         "-ab -sfoo -bc",
			Constructors: []fftest.Constructor{fftest.CoreConstructor},
			Args:         []string{`-ab`, `-sfoo`, `-bc`},
			Want:         fftest.Vars{A: true, B: true, C: true, S: "foo"},
		},
		{
			Name:         "-absfoo -c",
			Constructors: []fftest.Constructor{fftest.CoreConstructor},
			Args:         []string{`-absfoo`, `-c`},
			Want:         fftest.Vars{A: true, B: true, C: true, S: "foo"},
		},
		{
			Name:         "-acs foo -b",
			Constructors: []fftest.Constructor{fftest.CoreConstructor},
			Args:         []string{`-acs`, `foo`, `-b`},
			Want:         fftest.Vars{A: true, B: true, C: true, S: "foo"},
		},
		{
			Name:         "-a true -b false -c true",
			Constructors: []fftest.Constructor{fftest.CoreConstructor},
			Args:         []string{`-a`, `true`, `-b`, `false`, `-c`, `true`},
			Want:         fftest.Vars{A: true, Args: []string{`true`, `-b`, `false`, `-c`, `true`}},
		},
		{
			Name:         "--str foo -h",
			Constructors: []fftest.Constructor{fftest.CoreConstructor},
			Args:         []string{`--str`, `foo`, `-h`},
			Want:         fftest.Vars{S: "foo", WantParseErrorIs: ff.ErrHelp},
		},
		{
			Name:         "--str foo --help -b",
			Constructors: []fftest.Constructor{fftest.CoreConstructor},
			Args:         []string{`--str`, `foo`, `--help`, `-b`},
			Want:         fftest.Vars{S: "foo", B: false, WantParseErrorIs: ff.ErrHelp},
		},
		{
			Name:         "--str= -a",
			Constructors: []fftest.Constructor{fftest.CoreConstructor},
			Args:         []string{`--str=`, `-a`},
			Want:         fftest.Vars{S: "", A: true},
		},
		{
			Name:         "-s foo -f 1.23",
			Constructors: []fftest.Constructor{fftest.CoreConstructor},
			Args:         []string{"-s", "foo", "-f", "1.23"},
			Want:         fftest.Vars{S: "foo", F: 1.23},
		},
		{
			Name:         "-a true -b true",
			Constructors: []fftest.Constructor{fftest.CoreConstructor},
			Args:         []string{"-a", "true", "-b", "true"},
			Want:         fftest.Vars{A: true, B: false, Args: []string{"true", "-b", "true"}},
		},
		{
			Name:         "--aflag true --cflag true",
			Constructors: []fftest.Constructor{fftest.CoreConstructor},
			Args:         []string{"--aflag", "true", "--cflag", "true"},
			Want:         fftest.Vars{A: true, B: false, C: true, Args: []string{}},
		},
		{
			Name:         "-a --bflag=false",
			Constructors: []fftest.Constructor{fftest.CoreConstructor},
			Default:      fftest.Vars{A: true, B: true},
			Args:         []string{"-a", "--bflag=false"},
			Want:         fftest.Vars{A: true, B: false, C: false, Args: []string{}},
		},
		{
			Name:         "-a false",
			Constructors: []fftest.Constructor{fftest.CoreConstructor},
			Default:      fftest.Vars{A: true, B: true},
			Args:         []string{"-a", "false"},
			Want:         fftest.Vars{A: true, B: true, C: false, Args: []string{"false"}},
		},
	}

	testcases.Run(t)
}

func TestParse_StdFlagSetAdapter(t *testing.T) {
	t.Parallel()

	testcases := fftest.TestCases{
		{
			Name:         "-singledash space values",
			Constructors: []fftest.Constructor{fftest.StdConstructor},
			Args:         []string{"-s", "foo", "-f", "1.23"},
			Want:         fftest.Vars{S: "foo", F: 1.23},
		},
		{
			Name:         "-singledash space values bool",
			Constructors: []fftest.Constructor{fftest.StdConstructor},
			Args:         []string{"-a", "true", "-b", "true"},
			Want:         fftest.Vars{A: true, B: true},
		},
		{
			Name:         "--doubledash space values bool",
			Constructors: []fftest.Constructor{fftest.StdConstructor},
			Args:         []string{"--a", "true", "--c", "true"},
			Want:         fftest.Vars{A: true, C: true},
		},
		{
			Name:         "bool default true set false",
			Constructors: []fftest.Constructor{fftest.StdConstructor},
			Default:      fftest.Vars{A: true, B: true},
			Args:         []string{"-a", "-b=false"},
			Want:         fftest.Vars{A: true, B: false, C: false},
		},
		{
			Name:         "bool default true set false with spaces",
			Constructors: []fftest.Constructor{fftest.StdConstructor},
			Default:      fftest.Vars{A: true, B: true},
			Args:         []string{"-a", "false"},
			Want:         fftest.Vars{A: false, B: true, C: false},
		},
	}

	testcases.Run(t)
}

func TestParse_PlainParser(t *testing.T) {
	t.Parallel()

	testcases := fftest.TestCases{
		{
			Name:       "solo bool",
			ConfigFile: "testdata/solo_bool.conf",
			Want:       fftest.Vars{S: "x", B: true},
		},
		{
			Name:       "string with spaces",
			ConfigFile: "testdata/spaces.conf",
			Want:       fftest.Vars{S: "i am the very model of a modern major general"},
		},
		{
			Name:       "comments",
			ConfigFile: "testdata/comments.conf",
			Want: fftest.Vars{X: []string{
				"foo#bar",
				"foo# bar",
				"foo",
				"foo",
				`"foo#bar"#baz`,
				`"foo#bar"`,
				`"foo`,
			}},
		},
		{
			Name:       "newlines",
			ConfigFile: "testdata/newlines.conf",
			Want: fftest.Vars{X: []string{
				`hello\nworld\n`,
				`"hello\nworld\n"`,
			}},
		},
		{
			Name:       "WithConfigIgnoreUndefined not set",
			ConfigFile: "testdata/undefined.conf",
			Want:       fftest.Vars{WantParseErrorIs: ff.ErrUnknownFlag},
		},
		{
			Name:       "WithConfigIgnoreUndefined is set",
			ConfigFile: "testdata/undefined.conf",
			Options:    []ff.Option{ff.WithConfigIgnoreUndefinedFlags()},
			Want:       fftest.Vars{S: "one"},
		},
		{
			Name:       "WithFilesystem",
			ConfigFile: "testdata/1.conf",
			Options:    []ff.Option{ff.WithFilesystem(testdataConfigFS)},
			Want:       fftest.Vars{S: "bar", I: 99, B: true, D: 1 * time.Hour},
		},
	}

	testcases.Run(t)
}

func TestParse_types(t *testing.T) {
	t.Parallel()

	t.Run("ff.FlagSet", func(t *testing.T) {
		fs := ff.NewFlagSet(t.Name())
		foo := fs.String('f', "foo", "default-value", "foo string")
		if err := ff.Parse(fs, []string{"--foo=bar"}); err != nil {
			t.Fatal(err)
		}
		if want, have := "bar", *foo; want != have {
			t.Errorf("foo: want %q, have %q", want, have)
		}
	})

	t.Run("flag.FlagSet", func(t *testing.T) {
		fs := flag.NewFlagSet(t.Name(), flag.ContinueOnError)
		foo := fs.String("foo", "default-value", "foo string")
		if err := ff.Parse(fs, []string{"-foo", "bar"}); err != nil {
			t.Fatal(err)
		}
		if want, have := "bar", *foo; want != have {
			t.Errorf("foo: want %q, have %q", want, have)
		}
	})

	t.Run("invalid type", func(t *testing.T) {
		fs := "xxx" // should compile, but Parse should return an error
		if err := ff.Parse(fs, []string{"-foo", "bar"}); err == nil {
			t.Errorf("Parse(%T): want error, have none", fs)
		}
	})
}

func TestParse_stdfs(t *testing.T) {
	t.Parallel()

	configFile := filepath.Join(t.TempDir(), "config.conf")
	if err := os.WriteFile(configFile, []byte(`foo hello`), 0655); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	args := []string{"-config", configFile}

	fs := flag.NewFlagSet(t.Name(), flag.ContinueOnError)
	var (
		foo = fs.String("foo", "abc", "foo string")
		_   = fs.String("config", "", "config file")
	)

	if err := ff.Parse(fs, args,
		ff.WithConfigFileFlag("config"),
		ff.WithConfigFileParser(ff.PlainParser),
	); err != nil {
		t.Fatalf("Parse: %v", err)
	}

	if want, have := "hello", *foo; want != have {
		t.Errorf("foo: want %q, have %q", want, have)
	}
}
