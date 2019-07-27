package fftest

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"golang.org/x/xerrors"
)

// NewPair returns a predefined flag set, and a predefined set of variables that
// have been registered into it. Tests can call parse on the flag set with a
// variety of flags, config files, and env vars, and check the resulting effect
// on the vars.
func NewPair() (*flag.FlagSet, *Vars) {
	fs := flag.NewFlagSet("fftest", flag.ContinueOnError)

	var v Vars
	fs.StringVar(&v.S, "s", "", "string")
	fs.IntVar(&v.I, "i", 0, "int")
	fs.Float64Var(&v.F, "f", 0., "float64")
	fs.BoolVar(&v.B, "b", false, "bool")
	fs.DurationVar(&v.D, "d", time.Second, "time.Duration")

	return fs, &v
}

// Vars are a common set of variables used for testing.
type Vars struct {
	S string
	I int
	F float64
	B bool
	D time.Duration

	// ParseError is assigned as the result of Parse in test cases.
	ParseError error

	// If a test case expects an input to generate a parse error,
	// it can specify that error here. The Compare helper will
	// look for it using xerrors.Is.
	ParseErrorIs error

	// If a test case expects an input to generate a parse error,
	// it can specify part of that error string here. The Compare
	// helper will look for it using strings.Contains.
	ParseErrorString string
}

// Compare one set of vars with another
// and return an error on any difference.
func Compare(want, have *Vars) error {
	wantParseError := want.ParseErrorIs != nil || want.ParseErrorString != ""
	if wantParseError && have.ParseError == nil {
		return fmt.Errorf("want error (%v), have none", want)
	}
	if !wantParseError && have.ParseError != nil {
		return fmt.Errorf("want clean parse, have error (%v)", have)
	}

	if want.ParseErrorIs != nil && have.ParseError != nil && !xerrors.Is(have.ParseError, want.ParseErrorIs) {
		return fmt.Errorf("want wrapped error (%#+v), have error (%#+v)", want.ParseErrorIs, have.ParseError)
	}
	if want.ParseErrorString != "" && have.ParseError != nil && !strings.Contains(have.ParseError.Error(), want.ParseErrorString) {
		return fmt.Errorf("want error string (%q), have error (%v)", want.ParseErrorString, have.ParseError)
	}

	if want.S != have.S {
		return fmt.Errorf("S: want %q, have %q", want.S, have.S)
	}
	if want.I != have.I {
		return fmt.Errorf("I: want %d, have %d", want.I, have.I)
	}
	if want.F != have.F {
		return fmt.Errorf("F: want %f, have %f", want.F, have.F)
	}
	if want.B != have.B {
		return fmt.Errorf("B: want %v, have %v", want.B, have.B)
	}
	if want.D != have.D {
		return fmt.Errorf("D: want %s, have %s", want.D, have.D)
	}

	return nil
}
