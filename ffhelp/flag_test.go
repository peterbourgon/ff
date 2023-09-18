package ffhelp_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffhelp"
	"github.com/peterbourgon/ff/v4/ffval"
)

func TestFlagFormat(t *testing.T) {
	t.Parallel()

	fs := ff.NewFlagSet("flags")

	var foo ffhelp.Flag
	if f, err := fs.AddFlag(ff.FlagConfig{
		ShortName:   'f',
		LongName:    "foo",
		Value:       ffval.NewValueDefault(new(string), "hello world"),
		Usage:       "usage text",
		Placeholder: "STR",
	}); err != nil {
		t.Fatal(err)
	} else {
		foo = ffhelp.WrapFlag(f)
	}

	var bar ffhelp.Flag
	if f, err := fs.AddFlag(ff.FlagConfig{
		ShortName: 'b',
		Value:     ffval.NewValueDefault(new(bool), false),
	}); err != nil {
		t.Fatal(err)
	} else {
		bar = ffhelp.WrapFlag(f)
	}

	var baz ffhelp.Flag
	if f, err := fs.AddFlag(ff.FlagConfig{
		LongName:    "baz",
		Value:       ffval.NewValueDefault(new(bool), true),
		Placeholder: "XX",
	}); err != nil {
		t.Fatal(err)
	} else {
		baz = ffhelp.WrapFlag(f)
	}

	for _, testcase := range []struct {
		format, wantFoo, wantBar, wantBaz string
	}{
		{format: `%s`, wantFoo: `f, foo`, wantBar: `b`, wantBaz: `baz`},
		{format: `%+s`, wantFoo: `-f, --foo`, wantBar: `-b`, wantBaz: `--baz`},
		{format: `%#+s`, wantFoo: `-f, --foo`, wantBar: `-b`, wantBaz: `    --baz`},
		{format: `%v`, wantFoo: `f, foo STR`, wantBar: `b`, wantBaz: `baz XX`},
		{format: `%+v`, wantFoo: `-f, --foo STR`, wantBar: `-b`, wantBaz: `--baz XX`},
		{format: `%#+v`, wantFoo: `-f, --foo STR`, wantBar: `-b`, wantBaz: `    --baz XX`},
		{format: `%n`, wantFoo: `f`, wantBar: `b`, wantBaz: ``},
		{format: `%+n`, wantFoo: `-f`, wantBar: `-b`, wantBaz: ``},
		{format: `%l`, wantFoo: `foo`, wantBar: ``, wantBaz: `baz`},
		{format: `%+l`, wantFoo: `--foo`, wantBar: ``, wantBaz: `--baz`},
		{format: `%d`, wantFoo: `hello world`, wantBar: ``, wantBaz: `true`},
		{format: `%u`, wantFoo: `usage text`, wantBar: ``, wantBaz: ``},
		{format: `%k`, wantFoo: `STR`, wantBar: ``, wantBaz: `XX`},
	} {
		t.Run(testcase.format, func(t *testing.T) {
			if want, have := testcase.wantFoo, fmt.Sprintf(testcase.format, foo); want != have {
				t.Errorf("foo: want '%s', have '%s'", want, have)
			}
			if want, have := testcase.wantBar, fmt.Sprintf(testcase.format, bar); want != have {
				t.Errorf("bar: want '%s', have '%s'", want, have)
			}
			if want, have := testcase.wantBaz, fmt.Sprintf(testcase.format, baz); want != have {
				t.Errorf("baz: want '%s', have '%s'", want, have)
			}
		})
	}

	t.Run("Empties", func(t *testing.T) {
		fs := ff.NewFlagSet(t.Name())

		fooFlag, _ := fs.AddFlag(ff.FlagConfig{
			LongName:      "foo",
			Value:         new(ffval.Int),
			Usage:         "foo value",
			NoPlaceholder: true,
		})

		foo := ffhelp.WrapFlag(fooFlag)

		if want, have := "", fmt.Sprintf("%k", foo); want != have {
			t.Errorf("foo: Placeholder (%%k): want '%s', have '%s'", want, have)
		}

		if want, have := "0", fmt.Sprintf("%d", foo); want != have {
			t.Errorf("foo: Default: (%%d): want '%s', have '%s'", want, have)
		}

		barFlag, _ := fs.AddFlag(ff.FlagConfig{
			ShortName: 'b',
			LongName:  "bar",
			Value:     ffval.NewValueDefault(new(time.Duration), time.Second),
			Usage:     "bar value",
			NoDefault: true,
		})

		bar := ffhelp.WrapFlag(barFlag)

		if want, have := "DURATION", fmt.Sprintf("%k", bar); want != have {
			t.Errorf("bar: Placeholder (%%k): want '%s', have '%s'", want, have)
		}

		if want, have := "", fmt.Sprintf("%d", bar); want != have {
			t.Errorf("bar: Default: (%%d): want '%s', have '%s'", want, have)
		}
	})
}
