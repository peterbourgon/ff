package fftest

import (
	"errors"
	"flag"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
)

// Pair defines and returns an empty flag set and vars assigned to it.
func Pair() (*flag.FlagSet, *Vars) {
	fs := flag.NewFlagSet("fftest", flag.ContinueOnError)
	vars := DefaultVars(fs)
	return fs, vars
}

// DefaultVars registers a predefined set of variables to the flag set.
// Tests can call parse on the flag set with a variety of flags, config files,
// and env vars, and check the resulting effect on the vars.
func DefaultVars(fs *flag.FlagSet) *Vars {
	var v Vars
	fs.StringVar(&v.S, "s", "", "string")
	fs.IntVar(&v.I, "i", 0, "int")
	fs.Float64Var(&v.F, "f", 0., "float64")
	fs.BoolVar(&v.B, "b", false, "bool")
	fs.DurationVar(&v.D, "d", 0*time.Second, "time.Duration")
	fs.Var(&v.X, "x", "collection of strings (repeatable)")
	return &v
}

// NonzeroDefaultVars is like DefaultVars, but provides each primitive flag with
// a nonzero default value. This is useful for tests that explicitly provide a
// zero value for the type.
func NonzeroDefaultVars(fs *flag.FlagSet) *Vars {
	var v Vars
	fs.StringVar(&v.S, "s", "foo", "string")
	fs.IntVar(&v.I, "i", 123, "int")
	fs.Float64Var(&v.F, "f", 9.99, "float64")
	fs.BoolVar(&v.B, "b", true, "bool")
	fs.DurationVar(&v.D, "d", 3*time.Hour, "time.Duration")
	fs.Var(&v.X, "x", "collection of strings (repeatable)")
	return &v
}

// NestedDefaultVars is similar to DefaultVars, but uses nested flag names.
func NestedDefaultVars(delimiter string) func(fs *flag.FlagSet) *Vars {
	return func(fs *flag.FlagSet) *Vars {
		var v Vars
		fs.StringVar(&v.S, fmt.Sprintf("foo%ss", delimiter), "", "string")
		fs.IntVar(&v.I, fmt.Sprintf("bar%[1]snested%[1]si", delimiter), 0, "int")
		fs.Float64Var(&v.F, fmt.Sprintf("bar%[1]snested%[1]sf", delimiter), 0., "float64")
		fs.BoolVar(&v.B, fmt.Sprintf("foo%sb", delimiter), false, "bool")
		fs.Var(&v.X, fmt.Sprintf("baz%[1]snested%[1]sx", delimiter), "collection of strings (repeatable)")
		return &v
	}
}

// Vars are a common set of variables used for testing.
type Vars struct {
	S string
	I int
	F float64
	B bool
	D time.Duration
	X StringSlice

	// ParseError should be assigned as the result of Parse in tests.
	ParseError error

	// If a test case expects an input to generate a parse error,
	// it can specify that error here. The Compare helper will
	// look for it using errors.Is.
	WantParseErrorIs error

	// If a test case expects an input to generate a parse error,
	// it can specify part of that error string here. The Compare
	// helper will look for it using strings.Contains.
	WantParseErrorString string
}

// Compare one set of vars with another
// and t.Error on any difference.
func Compare(t *testing.T, want, have *Vars) {
	t.Helper()

	if want.WantParseErrorIs != nil || want.WantParseErrorString != "" {
		if want.WantParseErrorIs != nil && have.ParseError == nil {
			t.Errorf("want error (%v), have none", want.WantParseErrorIs)
		}
		if want.WantParseErrorString != "" && have.ParseError == nil {
			t.Errorf("want error (%q), have none", want.WantParseErrorString)
		}
		if want.WantParseErrorIs == nil && want.WantParseErrorString == "" && have.ParseError != nil {
			t.Errorf("want clean parse, have error (%v)", have.ParseError)
		}
		if want.WantParseErrorIs != nil && have.ParseError != nil && !errors.Is(have.ParseError, want.WantParseErrorIs) {
			t.Errorf("want wrapped error (%#+v), have error (%#+v)", want.WantParseErrorIs, have.ParseError)
		}
		if want.WantParseErrorString != "" && have.ParseError != nil && !strings.Contains(have.ParseError.Error(), want.WantParseErrorString) {
			t.Errorf("want error string (%q), have error (%v)", want.WantParseErrorString, have.ParseError)
		}
		return
	}

	if have.ParseError != nil {
		t.Errorf("error: %v", have.ParseError)
	}

	if want.S != have.S {
		t.Errorf("var S: want %q, have %q", want.S, have.S)
	}
	if want.I != have.I {
		t.Errorf("var I: want %d, have %d", want.I, have.I)
	}
	if want.F != have.F {
		t.Errorf("var F: want %f, have %f", want.F, have.F)
	}
	if want.B != have.B {
		t.Errorf("var B: want %v, have %v", want.B, have.B)
	}
	if want.D != have.D {
		t.Errorf("var D: want %s, have %s", want.D, have.D)
	}
	if !reflect.DeepEqual(want.X, have.X) {
		t.Errorf("var X: want %v, have %v", want.X, have.X)
	}
}

// StringSlice is a flag.Value that collects each Set string
// into a slice, allowing for repeated flags.
type StringSlice []string

// Set implements flag.Value and appends the string to the slice.
func (ss *StringSlice) Set(s string) error {
	(*ss) = append(*ss, s)
	return nil
}

// String implements flag.Value and returns the list of
// strings, or "..." if no strings have been added.
func (ss *StringSlice) String() string {
	if len(*ss) <= 0 {
		return "..."
	}
	return strings.Join(*ss, ", ")
}
