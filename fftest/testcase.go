package fftest

import (
	"os"
	"testing"

	"github.com/peterbourgon/ff/v4"
)

// TestCases are a collection of test cases that can be run as a group.
type TestCases []ParseTest

// Run the test cases in order.
func (tcs TestCases) Run(t *testing.T, options ...ff.Option) {
	t.Helper()
	for _, tc := range tcs {
		t.Run(tc.Name, func(t *testing.T) {
			t.Helper()
			tc.Run(t, options...)
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
func (tc *ParseTest) Run(t *testing.T, options ...ff.Option) {
	t.Helper()

	// The test case options are the most specific, and so the highest priority.
	opts := tc.Options

	// The options passed to run are the next-highest priority. Options are
	// evaluated first-to-last, and later options override earlier options, so
	// lower-priority options should come before higher-priority options.
	opts = append(options, opts...)

	// Default options have lowest priority, and so are first in the list.
	if tc.ConfigFile != "" {
		opts = append(
			[]ff.Option{ff.WithConfigFile(tc.ConfigFile), ff.WithConfigFileParser(ff.PlainParser)},
			opts...,
		)
	}

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
