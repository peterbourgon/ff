package ff_test

import (
	"errors"
	"flag"
	"strings"
	"testing"
	"time"

	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffhelp"
	"github.com/peterbourgon/ff/v4/fftest"
	"github.com/peterbourgon/ff/v4/ffval"
)

func TestCoreFlags_Basics(t *testing.T) {
	t.Parallel()

	for _, argstr := range []string{
		"",
		"-b",
		"-d 250ms",
		"-b -d250ms",
		"-bd250ms",
		"--duration 250ms --string=nondefault",
	} {
		t.Run(argstr, func(t *testing.T) {
			fs := ff.NewFlags("myset")
			fs.Bool('b', "boolean", false, "boolean flag")
			fs.StringLong("string", "default", "string flag")
			fs.Duration('d', "duration", 250*time.Millisecond, "duration flag")
			fftest.TestFlags(t, fs, strings.Fields(argstr))
		})
	}
}

func TestStdFlags_Basics(t *testing.T) {
	t.Parallel()

	for _, argstr := range []string{
		"",
		"-b",
		"-d=250ms",
		"-string 250ms",
		"--string=250ms",
		"--string 250ms",
	} {
		t.Run(argstr, func(t *testing.T) {
			stdfs := flag.NewFlagSet("myset", flag.ContinueOnError)
			stdfs.Bool("b", false, "boolean flag")
			stdfs.String("string", "default", "string flag")
			stdfs.Duration("d", 250*time.Millisecond, "duration flag")
			corefs := ff.NewStdFlags(stdfs)
			fftest.TestFlags(t, corefs, strings.Fields(argstr))
		})
	}
}

func TestCoreFlags_Bool(t *testing.T) {
	t.Parallel()

	t.Run("add bool flag", func(t *testing.T) {
		var (
			fs     = ff.NewFlags(t.Name())
			bflag  bool
			bvalue = ffval.NewValueDefault(&bflag, true)
		)

		if _, err := fs.AddFlag(ff.CoreFlagConfig{
			ShortName: 'b',
			Value:     bvalue,
		}); err == nil {
			t.Errorf("add default true bool with no long name: want error, have none")
		}

		if _, err := fs.AddFlag(ff.CoreFlagConfig{
			ShortName: 'b',
			LongName:  "bflag",
			Value:     bvalue,
		}); err != nil {
			t.Errorf("add default true bool with long name: %v", err)
		}
	})

	for _, test := range []struct {
		args    []string
		wantX   bool
		wantY   bool
		wantErr error
	}{
		{args: []string{"--xflag"}, wantX: true, wantY: true},
		{args: []string{"--xflag=true"}, wantX: true, wantY: true},
		{args: []string{"--xflag", "true"}, wantX: true, wantY: true},
		{args: []string{"-x=true"}, wantX: false, wantY: true, wantErr: ff.ErrUnknownFlag}, // = interpreted as flag
		{args: []string{"-x"}, wantX: true, wantY: true},
		{args: []string{"-x", "false"}, wantX: true, wantY: true}, // false interpreted as argument
		{args: []string{"-y"}, wantX: false, wantY: true},
		{args: []string{"--yflag=false"}, wantX: false, wantY: false},
		{args: []string{"--yflag", "false"}, wantX: false, wantY: false},
		{args: []string{"--yflag", "false", "-y"}, wantX: false, wantY: true},
		{args: []string{"-y=false"}, wantX: false, wantY: false, wantErr: ff.ErrUnknownFlag}, // = interpreted as flag
		{args: []string{"-h"}, wantX: false, wantY: true, wantErr: ff.ErrHelp},
		{args: []string{"--help"}, wantX: false, wantY: true, wantErr: ff.ErrHelp},
		{args: []string{"--xflag", "-h"}, wantX: true, wantY: true, wantErr: ff.ErrHelp},
		{args: []string{"-y", "--help"}, wantX: false, wantY: false, wantErr: ff.ErrHelp},
	} {
		t.Run(strings.Join(test.args, " "), func(t *testing.T) {
			fs := ff.NewFlags(t.Name())
			xflag := fs.Bool('x', "xflag", false, "one boolean flag")
			yflag := fs.Bool('y', "yflag", true, "another boolean flag")
			err := fs.Parse(test.args)
			switch {
			case test.wantErr == nil && err == nil:
				break // good, and we should test the other stuff
			case test.wantErr == nil && err != nil:
				t.Fatalf("want no error, got error (%v)", err)
			case test.wantErr != nil && err == nil:
				t.Fatalf("want error (%v), got none", test.wantErr)
			case test.wantErr != nil && err != nil && !errors.Is(err, test.wantErr):
				t.Fatalf("want error (%v), got different error (%v)", test.wantErr, err)
			case test.wantErr != nil && err != nil && errors.Is(err, test.wantErr):
				return // good, but we shouldn't test anything else
			}
			if want, have := test.wantX, *xflag; want != have {
				t.Errorf("x: want %v, have %v", want, have)
			}
			if want, have := test.wantY, *yflag; want != have {
				t.Errorf("y: want %v, have %v", want, have)
			}
		})
	}
}

func TestStdFlags_Bool(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		args    []string
		wantX   bool
		wantY   bool
		wantErr error
	}{
		{args: []string{"-xflag"}, wantX: true, wantY: true},
		{args: []string{"-xflag=true"}, wantX: true, wantY: true},
		{args: []string{"-xflag", "true"}, wantX: true, wantY: true},
		{args: []string{"--xflag", "true"}, wantX: true, wantY: true},
		{args: []string{"--xflag=true"}, wantX: true, wantY: true},
		{args: []string{"-y"}, wantX: false, wantY: true},
		{args: []string{"-y=false"}, wantX: false, wantY: false},
		{args: []string{"-y", "false"}, wantX: false, wantY: false},
		{args: []string{"--y=false"}, wantX: false, wantY: false},
		{args: []string{"--y", "false"}, wantX: false, wantY: false},
		{args: []string{"--y", "false", "-y"}, wantX: false, wantY: true},
		{args: []string{"-h"}, wantX: false, wantY: true, wantErr: ff.ErrHelp},
		{args: []string{"--help"}, wantX: false, wantY: true, wantErr: ff.ErrHelp},
		{args: []string{"-xflag", "-h"}, wantX: true, wantY: true, wantErr: ff.ErrHelp},
		{args: []string{"--y=false", "--help"}, wantX: false, wantY: false, wantErr: ff.ErrHelp},
	} {
		t.Run(strings.Join(test.args, " "), func(t *testing.T) {
			stdfs := flag.NewFlagSet(t.Name(), flag.ContinueOnError)
			xflag := stdfs.Bool("xflag", false, "one boolean flag")
			yflag := stdfs.Bool("y", true, "another boolean flag")
			corefs := ff.NewStdFlags(stdfs)
			err := corefs.Parse(test.args)
			switch {
			case test.wantErr == nil && err == nil:
				break // good, and we should test the other stuff
			case test.wantErr == nil && err != nil:
				t.Fatalf("want no error, got error (%v)", err)
			case test.wantErr != nil && err == nil:
				t.Fatalf("want error (%v), got none", test.wantErr)
			case test.wantErr != nil && err != nil && !errors.Is(err, test.wantErr):
				t.Fatalf("want error (%v), got different error (%v)", test.wantErr, err)
			case test.wantErr != nil && err != nil && errors.Is(err, test.wantErr):
				return // good, but we shouldn't test anything else
			}
			if want, have := test.wantX, *xflag; want != have {
				t.Errorf("x: want %v, have %v", want, have)
			}
			if want, have := test.wantY, *yflag; want != have {
				t.Errorf("y: want %v, have %v", want, have)
			}
		})
	}
}

func TestCoreFlags_HelpFlag(t *testing.T) {
	t.Parallel()

	fs := ff.NewFlags(t.Name())
	helpflag := fs.BoolLong("help", false, "alternative help flag")

	// -h should still trigger ErrHelp.
	if err := fs.Parse([]string{"-h"}); !errors.Is(err, ff.ErrHelp) {
		t.Errorf("Parse(-h): want %v, have %v", ff.ErrHelp, err)
	}

	if err := fs.Reset(); err != nil {
		t.Fatalf("Reset(): %v", err)
	}

	// --help should not.
	if err := fs.Parse([]string{"--help"}); err != nil {
		t.Errorf("Parse(--help): error: %v", err)
	}

	// It should set the flag we defined.
	if want, have := true, *helpflag; want != have {
		t.Errorf("h: want %v, have %v", want, have)
	}
}

func TestCoreFlags_GetFlag(t *testing.T) {
	t.Parallel()

	fs := ff.NewFlags(t.Name())
	fs.IntLong("foo", 0, "first flag")
	fs.IntShort('f', 0, "second flag")

	f1, ok := fs.GetFlag("foo")
	if !ok {
		t.Fatalf(`GetFlag("foo"): returned not-OK`)
	}
	if want, have := "first flag", f1.GetUsage(); want != have {
		t.Errorf(`GetFlag("foo"): GetUsage: want %q, have %q`, want, have)
	}

	f2, ok := fs.GetFlag("f")
	if !ok {
		t.Fatalf(`GetFlag("f"): returned not-OK`)
	}
	if want, have := "second flag", f2.GetUsage(); want != have {
		t.Errorf(`GetFlag("f"): GetUsage: want %q, have %q`, want, have)
	}
}

func TestCoreFlags_NoDefault(t *testing.T) {
	t.Parallel()

	fs := ff.NewFlags(t.Name())
	alpha, _ := fs.AddFlag(ff.CoreFlagConfig{LongName: "alpha", Value: &ffval.Duration{}, Usage: "zero duration"})
	beta, _ := fs.AddFlag(ff.CoreFlagConfig{LongName: "beta", Value: &ffval.Duration{}, Usage: "zero duration with NoDefault", NoDefault: true})

	if want, have := "0s", alpha.GetDefault(); want != have {
		t.Errorf("alpha: default: want %q, have %q", want, have)
	}

	if want, have := "", beta.GetDefault(); want != have {
		t.Errorf("beta: default: want %q, have %q", want, have)
	}
}

func TestCoreFlags_NoPlaceholder(t *testing.T) {
	t.Parallel()

	fs := ff.NewFlags(t.Name())
	alpha, _ := fs.AddFlag(ff.CoreFlagConfig{LongName: "alpha", Value: &ffval.Bool{}, Usage: "alpha", NoPlaceholder: true})
	beta, _ := fs.AddFlag(ff.CoreFlagConfig{LongName: "beta", Value: ffval.NewValueDefault(new(bool), true), Usage: "beta", NoPlaceholder: true})
	delta, _ := fs.AddFlag(ff.CoreFlagConfig{LongName: "delta", Value: ffval.NewValueDefault(new(bool), true), Usage: "delta `D` flag", NoPlaceholder: true})
	kappa, _ := fs.AddFlag(ff.CoreFlagConfig{LongName: "kappa", Value: ffval.NewValue(new(bool)), Usage: "kappa `K` flag", NoPlaceholder: true})

	for _, f := range []ff.Flag{alpha, beta, delta, kappa} {
		if want, have := "", f.GetPlaceholder(); want != have {
			t.Errorf("%s: want %q, have %q", ffhelp.WrapFlag(f), want, have)
		}
	}
}

func TestCoreFlags_Get(t *testing.T) {
	t.Parallel()

	fs := ff.NewFlags(t.Name())
	f, err := fs.AddFlag(ff.CoreFlagConfig{
		LongName:    "alpha",
		Value:       new(ffval.Int),
		Placeholder: "X",
	})
	if err != nil {
		t.Fatal(err)
	}

	if want, have := "0", f.GetDefault(); want != have {
		t.Errorf("GetDefault: want %q, have %q", want, have)
	}

	if want, have := "X", f.GetPlaceholder(); want != have {
		t.Errorf("GetPlaceholder: want %q, have %q", want, have)
	}
}
