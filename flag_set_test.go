package ff_test

import (
	"errors"
	"flag"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffhelp"
	"github.com/peterbourgon/ff/v4/fftest"
	"github.com/peterbourgon/ff/v4/ffval"
)

func TestFlagSet_Basics(t *testing.T) {
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
			fs := ff.NewFlagSet("myset")
			fs.Bool('b', "boolean", "boolean flag")
			fs.StringLong("string", "default", "string flag")
			fs.Duration('d', "duration", 250*time.Millisecond, "duration flag")
			fftest.ValidateFlags(t, fs, strings.Fields(argstr))
		})
	}
}

func TestFlagSet_Bool(t *testing.T) {
	t.Parallel()

	t.Run("add bool flag", func(t *testing.T) {
		var (
			fs     = ff.NewFlagSet(t.Name())
			bflag  bool
			bvalue = ffval.NewValueDefault(&bflag, true)
		)

		if _, err := fs.AddFlag(ff.FlagConfig{
			ShortName: 'b',
			Value:     bvalue,
		}); err == nil {
			t.Errorf("add default true bool with no long name: want error, have none")
		}

		if _, err := fs.AddFlag(ff.FlagConfig{
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
			fs := ff.NewFlagSet(t.Name())
			xflag := fs.Bool('x', "xflag", "one boolean flag")
			yflag := fs.BoolDefault('y', "yflag", true, "another boolean flag")
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
			err := ff.Parse(stdfs, test.args)
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

func TestFlagSet_HelpFlag(t *testing.T) {
	t.Parallel()

	fs := ff.NewFlagSet(t.Name())
	helpflag := fs.BoolLong("help", "alternative help flag")

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

func TestFlagSet_GetFlag(t *testing.T) {
	t.Parallel()

	fs := ff.NewFlagSet(t.Name())
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

func TestFlagSet_NoDefault(t *testing.T) {
	t.Parallel()

	fs := ff.NewFlagSet(t.Name())
	alpha, _ := fs.AddFlag(ff.FlagConfig{LongName: "alpha", Value: &ffval.Duration{}, Usage: "zero duration"})
	beta, _ := fs.AddFlag(ff.FlagConfig{LongName: "beta", Value: &ffval.Duration{}, Usage: "zero duration with NoDefault", NoDefault: true})

	if want, have := "0s", alpha.GetDefault(); want != have {
		t.Errorf("alpha: default: want %q, have %q", want, have)
	}

	if want, have := "", beta.GetDefault(); want != have {
		t.Errorf("beta: default: want %q, have %q", want, have)
	}
}

func TestFlagSet_NoPlaceholder(t *testing.T) {
	t.Parallel()

	fs := ff.NewFlagSet(t.Name())
	alpha, _ := fs.AddFlag(ff.FlagConfig{LongName: "alpha", Value: &ffval.Bool{}, Usage: "alpha", NoPlaceholder: true})
	beta, _ := fs.AddFlag(ff.FlagConfig{LongName: "beta", Value: ffval.NewValueDefault(new(bool), true), Usage: "beta", NoPlaceholder: true})
	delta, _ := fs.AddFlag(ff.FlagConfig{LongName: "delta", Value: ffval.NewValueDefault(new(bool), true), Usage: "delta `D` flag", NoPlaceholder: true})
	kappa, _ := fs.AddFlag(ff.FlagConfig{LongName: "kappa", Value: ffval.NewValue(new(bool)), Usage: "kappa `K` flag", NoPlaceholder: true})

	for _, f := range []ff.Flag{alpha, beta, delta, kappa} {
		if want, have := "", f.GetPlaceholder(); want != have {
			t.Errorf("%s: want %q, have %q", ffhelp.WrapFlag(f), want, have)
		}
	}
}

func TestFlagSet_Get(t *testing.T) {
	t.Parallel()

	fs := ff.NewFlagSet(t.Name())
	f, err := fs.AddFlag(ff.FlagConfig{
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

func TestFlagSet_invalid(t *testing.T) {
	t.Parallel()

	t.Run("same short and long name", func(t *testing.T) {
		defer func() {
			if x := recover(); x == nil {
				t.Errorf("want panic, have none")
			} else {
				t.Logf("have expected panic (%v)", x)
			}
		}()
		fs := ff.NewFlagSet(t.Name())
		fs.Bool('b', "b", "this should panic")
	})

	t.Run("duplicate short name", func(t *testing.T) {
		defer func() {
			if x := recover(); x == nil {
				t.Errorf("want panic, have none")
			} else {
				t.Logf("have expected panic (%v)", x)
			}
		}()
		fs := ff.NewFlagSet(t.Name())
		_ = fs.Bool('a', "alpha", "this should be OK")
		_ = fs.Bool('a', "apple", "this should panic")
	})

	t.Run("same fold short name", func(t *testing.T) {
		defer func() {
			if x := recover(); x == nil {
				t.Logf("have not expected panic (%v)", x)
			} else {
				t.Errorf("want peace, have panic")
			}
		}()
		fs := ff.NewFlagSet(t.Name())
		_ = fs.Bool('a', "alpha", "this should be OK")
		_ = fs.Bool('A', "apple", "this should be OK")
	})

	t.Run("duplicate long name", func(t *testing.T) {
		defer func() {
			if x := recover(); x == nil {
				t.Errorf("want panic, have none")
			} else {
				t.Logf("have expected panic (%v)", x)
			}
		}()
		fs := ff.NewFlagSet(t.Name())
		_ = fs.Bool('a', "alpha", "this should be OK")
		_ = fs.Bool('b', "alpha", "this should panic")
	})
}

func TestFlagSet_structs(t *testing.T) {
	t.Parallel()

	type myFlags struct {
		Alpha string `ff:"short: a, long: alpha, default: alpha-default, usage: alpha string"`
		Beta  int    `ff:"          long: beta,  placeholder: β,         usage: beta int"`
		Delta bool   `ff:"short: d,              nodefault,              usage: delta bool"`

		Epsilon bool    `ff:"| short=e | long=epsilon | nodefault    | usage: epsilon bool          |"`
		Gamma   string  `ff:"| short=g | long=gamma   |              | usage: 'usage, with a comma' |"`
		Iota    float64 `ff:"|         | long=iota    | default=0.43 | usage: iota float            |"`
	}

	var flags myFlags
	fs := ff.NewFlagSetFrom(t.Name(), &flags)

	if want, have := fftest.UnindentString(`
		NAME
		  TestFlagSet_structs

		FLAGS
		  -a, --alpha STRING   alpha string (default: alpha-default)
		      --beta β         beta int (default: 0)
		  -d                   delta bool
		  -e, --epsilon        epsilon bool
		  -g, --gamma STRING   usage, with a comma
		      --iota FLOAT64   iota float (default: 0.43)
	`), fftest.UnindentString(ffhelp.Flags(fs).String()); want != have {
		t.Error(fftest.DiffString(want, have))
	}

	for _, testcase := range []struct {
		args string
		want myFlags
	}{
		{
			args: "--alpha=x",
			want: myFlags{Alpha: "x", Iota: 0.43},
		},
		{
			args: "-e --iota=1.23",
			want: myFlags{Alpha: "alpha-default", Epsilon: true, Iota: 1.23},
		},
		{
			args: "-gabc -d",
			want: myFlags{Alpha: "alpha-default", Delta: true, Gamma: "abc", Iota: 0.43},
		},
	} {
		t.Run(testcase.args, func(t *testing.T) {
			if err := fs.Reset(); err != nil {
				t.Fatalf("Reset: %v", err)
			}
			if err := ff.Parse(fs, strings.Fields(testcase.args)); err != nil {
				t.Fatalf("Parse: %v", err)
			}
			if want, have := testcase.want, flags; !reflect.DeepEqual(want, have) {
				t.Errorf("\nwant %+#v\nhave %#+v", want, have)
			}
		})
	}

	{
		if err := fs.Reset(); err != nil {
			t.Fatalf("Reset: %v", err)
		}
		if err := ff.Parse(fs, []string{}); err != nil {
			t.Fatalf("Parse: %v", err)
		}
		if want, have := "alpha-default", flags.Alpha; want != have {
			t.Errorf("alpha: want %q, have %q", want, have)
		}
		if want, have := 0, flags.Beta; want != have {
			t.Errorf("beta: want %v, have %v", want, have)
		}
		if want, have := false, flags.Delta; want != have {
			t.Errorf("delta: want %v, have %v", want, have)
		}
	}

	{
		if err := fs.Reset(); err != nil {
			t.Fatalf("Reset: %v", err)
		}
		if err := ff.Parse(fs, []string{"-afoo", "--beta", "7", "-d"}); err != nil {
			t.Fatalf("Parse: %v", err)
		}
		if want, have := "foo", flags.Alpha; want != have {
			t.Errorf("alpha: want %q, have %q", want, have)
		}
		if want, have := 7, flags.Beta; want != have {
			t.Errorf("beta: want %v, have %v", want, have)
		}
		if want, have := true, flags.Delta; want != have {
			t.Errorf("delta: want %v, have %v", want, have)
		}
	}

	t.Run("implements", func(t *testing.T) {
		var flags struct {
			Foo ffval.UniqueList[string] `ff:"longname=foo , usage=foo strings"`
			Bar ffval.Value[int]         `ff:"longname=bar , usage=bar int"`
		}

		fs := ff.NewFlagSet(t.Name())
		if err := fs.AddStruct(&flags); err != nil { // should allow
			t.Fatalf("AddStruct: %v", err)
		}

		if err := ff.Parse(fs, []string{"--foo=a", "--foo", "b"}); err != nil {
			t.Fatalf("Parse: %v", err)
		}

		if want, have := []string{"a", "b"}, flags.Foo.Get(); !reflect.DeepEqual(want, have) {
			t.Errorf("foo: want %#+v, have %#+v", want, have)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		for i, st := range []any{
			&struct {
				A int `ff:"x"` // invalid tag data key
			}{},
			&struct {
				B int `ff:"short = a, longname=, usage=some usage"` // invalid long name
			}{},
			&struct {
				C int `ff:"short = ,"` // invalid short name
			}{},
			&struct {
				D *testing.T `ff:"long=alpha"` // invalid field type
			}{},
			&struct {
				E bool `ff:"s=e,l=e"` // identical short and long names
			}{},
			&struct {
				F string `ff:"long:' usage='value,u=this is a weird one"` // exercises long name validity
			}{},
			&struct {
				G string `ff:"long:'  '"` // value should be trimmed of spaces and therefore invalid
			}{},
		} {
			t.Run(fmt.Sprint(i+1), func(t *testing.T) {
				fs := ff.NewFlagSet(t.Name())
				if err := fs.AddStruct(st); err == nil {
					t.Errorf("want error, have none\n%s", ffhelp.Flags(fs))
				} else {
					t.Logf("have expected error (%v)", err)
				}
			})
		}
	})

	t.Run("dupe", func(t *testing.T) {
		fs := ff.NewFlagSet(t.Name())
		fs.Bool('a', "alpha", "some bool flag")

		var s struct {
			Apple string `ff:"short=a, long=apple"`
		}
		if err := fs.AddStruct(&s); err == nil {
			t.Errorf("want error, have none")
		} else {
			t.Logf("have expected error (%v)", err)
		}
	})
}

func TestFlagSet_StructIgnoreReset(t *testing.T) {
	t.Parallel()

	type A struct {
		Foo string `ff:"long=foo, usage=foo string, default=xxx" json:"foo"`
		Bar string `json:"bar"`
		Baz string `ff:"long=baz, usage=baz string"`
		Qux string
	}

	var aval A

	fs := ff.NewFlagSet(t.Name())
	if err := fs.AddStruct(&aval); err != nil {
		t.Fatalf("AddStruct(&aval): %v", err)
	}

	{
		args := []string{"--foo=abc", "--bar=def", "--baz=ghi"}
		if err := fs.Parse(args); err == nil {
			t.Errorf("ff.Parse(%v): want error, have none", args)
		}
	}

	{
		args := []string{"--foo=1", "--baz=2"}
		if err := fs.Parse(args); err != nil {
			t.Errorf("ff.Parse(%v): %v", args, err)
		}
		if want, have := "1", aval.Foo; want != have {
			t.Errorf("Foo: want %q, have %q", want, have)
		}
		if want, have := "2", aval.Baz; want != have {
			t.Errorf("Baz: want %q, have %q", want, have)
		}
	}

	{
		if err := fs.Reset(); err != nil {
			t.Fatalf("Reset: %v", err)
		}
		if want, have := "xxx", aval.Foo; want != have {
			t.Errorf("Foo: want %q, have %q", want, have)
		}
		if want, have := "", aval.Baz; want != have {
			t.Errorf("Baz: want %q, have %q", want, have)
		}
	}
}

func TestFlagSet_StructEmbedded(t *testing.T) {
	t.Parallel()

	type A struct {
		Foo string `ff:"short=f, long=foo, usage=foo string"`
		Bar int    `ff:"         long=bar, usage=bar int, default=32"`
	}

	type B struct {
		A
		Quux bool `ff:"short=q, long=quux, usage=quux bool"`
	}

	type C struct {
		*A
		Zombo bool `ff:"short=z, long=zombo, usage=zombo bool"`
	}

	fs := ff.NewFlagSet(t.Name())

	var aflags A
	if err := fs.AddStruct(&aflags); err != nil {
		t.Fatalf("AddStruct(&aflags): %v", err)
	}

	var bflags B
	if err := fs.AddStruct(&bflags); err != nil { // should not try to re-add flags in embedded A
		t.Fatalf("AddStruct(&bflags): %v", err)
	}

	var cflags C
	if err := fs.AddStruct(&cflags); err != nil { // should not try to re-add flags in embedded *A
		t.Fatalf("AddStruct(&cflags): %v", err)
	}

	var flagNames []string
	fs.WalkFlags(func(f ff.Flag) error {
		flagNames = append(flagNames, ffhelp.FormatFlag(f, "%+s"))
		return nil
	})
	if want, have := []string{"-f, --foo", "--bar", "-q, --quux", "-z, --zombo"}, flagNames; !reflect.DeepEqual(want, have) {
		t.Errorf("flag names: want %v, have %v", want, have)
	}
}

func TestFlagSet_Std(t *testing.T) {
	t.Parallel()

	stdfs := flag.NewFlagSet(t.Name(), flag.ContinueOnError)
	var (
		_ = stdfs.String("foo", "hello world", "foo string")
		_ = stdfs.Int("b", 123, "b int")
	)

	fs := ff.NewFlagSetFrom(stdfs.Name(), stdfs)

	var flagNames []string
	fs.WalkFlags(func(f ff.Flag) error {
		flagNames = append(flagNames, ffhelp.FormatFlag(f, "%+s"))
		return nil
	})

	// flag.FlagSet sorts flags lexicographically
	if want, have := []string{"--b", "--foo"}, flagNames; !reflect.DeepEqual(want, have) {
		t.Errorf("flag names: want %v, have %v", want, have)
	}
}
