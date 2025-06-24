package fftest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffenv"
	"github.com/peterbourgon/ff/v4/ffjson"
	"github.com/peterbourgon/ff/v4/fftoml"
	"github.com/peterbourgon/ff/v4/ffyaml"
)

// TestCases are a collection of test cases that can be run as a group.
type TestCases []ParseTest

// Run the test cases in order.
func (tcs TestCases) Run(t *testing.T) {
	t.Helper()
	for _, tc := range tcs {
		t.Run(tc.Name, func(t *testing.T) {
			t.Helper()
			tc.Run(t)
		})
	}
}

// ParseTest describes a parsing test scenario.
type ParseTest struct {
	Name         string
	Constructors []Constructor
	Default      Vars
	ConfigFile   string
	Environment  map[string]string
	Args         []string
	Options      []ff.Option
	Want         Vars
}

// Run the test case.
func (tc *ParseTest) Run(t *testing.T) {
	t.Helper()

	// Set up the options we'll pass to parse.
	var opts []ff.Option

	// Some default options.
	if tc.ConfigFile != "" {
		// Try to deduce a default parser from the config file.
		var parseFunc ff.ConfigFileParseFunc
		switch strings.ToLower(filepath.Ext(tc.ConfigFile)) {
		case ".json":
			parseFunc = ffjson.Parse
		case ".yaml":
			parseFunc = ffyaml.Parse
		case ".toml":
			parseFunc = fftoml.Parse
		case ".env":
			parseFunc = ffenv.Parse
			opts = append(opts, ff.WithEnvVars())
		default:
			parseFunc = ff.PlainParser
		}
		opts = append(opts, ff.WithConfigFile(tc.ConfigFile), ff.WithConfigFileParser(parseFunc))
	}

	// Any options in the test case.
	//
	// Options are evaluated first-to-last, and later options override earlier
	// ones, so higher-priority stuff should come after lower-priority stuff.
	opts = append(opts, tc.Options...)

	// If there are any environment variables, set them before running the
	// tests, and reset them afterwards. Note that this means test cases cannot
	// be run in parallel.
	if len(tc.Environment) > 0 {
		for k, v := range tc.Environment {
			defer os.Setenv(k, os.Getenv(k))
			os.Setenv(k, v)
		}
	}

	// If no constructors were explicitly specified, use the defaults.
	if len(tc.Constructors) <= 0 {
		tc.Constructors = DefaultConstructors
	}

	// Run the test case for each constructor.
	for _, constr := range tc.Constructors {
		t.Run(constr.Name, func(t *testing.T) {
			t.Helper()
			fs, vars := constr.Make(tc.Default)
			vars.ParseError = ff.Parse(fs, tc.Args, opts...)
			vars.Args = fs.GetArgs()
			Compare(t, &tc.Want, vars)
		})
	}
}
