package ff_test

import (
	"errors"
	"flag"
	"strings"
	"testing"
	"time"

	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/fftest"
)

func TestCoreFlagsBasics(t *testing.T) {
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

func TestCoreFlagsBool(t *testing.T) {
	t.Parallel()

	var (
		fs     = ff.NewFlags(t.Name())
		bflag  bool
		bvalue = ff.MakeFlagValue(true, &bflag)
	)

	if err := fs.AddFlag(ff.CoreFlagConfig{
		ShortName: 'b',
		Value:     bvalue,
	}); err == nil {
		t.Errorf("add default true bool with no long name: want error, have none")
	}

	if err := fs.AddFlag(ff.CoreFlagConfig{
		ShortName: 'b',
		LongName:  "bflag",
		Value:     bvalue,
	}); err != nil {
		t.Errorf("add default true bool with long name: %v", err)
	}
}

func TestCoreFlagsHelpFlag(t *testing.T) {
	t.Parallel()

	fs := ff.NewFlags(t.Name())
	helpflag := fs.BoolLong("help", false, "alternative help flag")

	// -h should still trigger ErrHelp.
	if err := fs.Parse([]string{"-h"}); !errors.Is(err, ff.ErrHelp) {
		t.Errorf("Parse(-h): want %v, have %v", ff.ErrHelp, err)
	}

	fs.Reset()

	// --help should not.
	if err := fs.Parse([]string{"--help"}); err != nil {
		t.Errorf("Parse(--help): error: %v", err)
	}

	// It should set the flag we defined.
	if want, have := true, *helpflag; want != have {
		t.Errorf("h: want %v, have %v", want, have)
	}
}

func TestStdFlagsBasics(t *testing.T) {
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

func TestStdFlagsBool(t *testing.T) {
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
