package fftest

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
)

// Vars are a common set of variables used for testing.
type Vars struct {
	S       string        // flag name `s`
	I       int           // flag name `i`
	F       float64       // flag name `f`
	A, B, C bool          // flag name `a`, `b`, `c`
	D       time.Duration // flag name `d`
	X       []string      // flag name `x`

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

	// Args left over after a successful parse.
	Args []string
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
		t.Errorf("parse error: %v", have.ParseError)
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
	if want.A != have.A {
		t.Errorf("var A: want %v, have %v", want.A, have.A)
	}
	if want.B != have.B {
		t.Errorf("var B: want %v, have %v", want.B, have.B)
	}
	if want.C != have.C {
		t.Errorf("var C: want %v, have %v", want.C, have.C)
	}
	if want.D != have.D {
		t.Errorf("var D: want %s, have %s", want.D, have.D)
	}
	if !reflect.DeepEqual(want.X, have.X) {
		t.Errorf("var X: want %v, have %v", want.X, have.X)
	}

	if len(want.Args) > 0 {
		if !reflect.DeepEqual(want.Args, have.Args) {
			t.Errorf("post-parse args: want %v, have %v", want.Args, have.Args)
		}
	}
}
