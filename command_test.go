package ff_test

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/peterbourgon/ff/v4"
)

func TestCommandNoFlagSet(t *testing.T) {
	t.Parallel()

	var (
		cmd = &ff.Command{Name: "root"}
		ctx = context.Background()
	)
	if err := cmd.ParseAndRun(ctx, []string{"-h"}); !errors.Is(err, ff.ErrHelp) {
		t.Errorf("err: want %v, have %v", ff.ErrHelp, err)
	}
	if err := cmd.ParseAndRun(ctx, []string{"--help"}); !errors.Is(err, ff.ErrHelp) {
		t.Errorf("err: want %v, have %v", ff.ErrHelp, err)
	}
	if err := cmd.ParseAndRun(ctx, []string{}); !errors.Is(err, ff.ErrNoExec) {
		t.Errorf("err: want %v, have %v", ff.ErrNoExec, err)
	}
}

func TestCommandReset(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	rootcmd, testvars := makeTestCommand(t)
	defaults := *testvars

	t.Run("first run", func(t *testing.T) {
		args := []string{"--verbose", "foo", "-b", "hello world"}
		if err := rootcmd.ParseAndRun(ctx, args); err != nil {
			t.Fatalf("first run: %v", err)
		}

		want := defaults // copy
		want.Verbose = true
		want.Beta = true

		compareTestCommandVars(t, want, *testvars)
	})

	t.Run("second run without reset", func(t *testing.T) {
		want := ff.ErrAlreadyParsed
		have := rootcmd.ParseAndRun(ctx, nil)
		if !errors.Is(have, want) {
			t.Errorf("second run without reset: want error %v, have %v", want, have)
		}
	})

	t.Run("reset", func(t *testing.T) {
		if err := rootcmd.Reset(); err != nil {
			t.Fatalf("reset: %s: %v", rootcmd.Name, err)
		}
	})

	t.Run("second run after reset", func(t *testing.T) {
		args := []string{"--config-file=my.conf", "foo", "bar", "-a3", "hello world"}
		if err := rootcmd.ParseAndRun(ctx, args); err != nil {
			t.Fatalf("second run: %v", err)
		}

		want := defaults // copy
		want.ConfigFile = "my.conf"
		want.Alpha = 3

		compareTestCommandVars(t, want, *testvars)
	})
}

func makeTestCommand(t *testing.T) (*ff.Command, *testCommandVars) {
	t.Helper()

	var vars testCommandVars

	rootFlagSet := ff.NewSet("root")
	rootFlagSet.BoolVar(&vars.Verbose, 'v', "verbose", false, "verbose logging")
	rootFlagSet.StringVar(&vars.ConfigFile, 0, "config-file", "", "config file")
	rootCommand := &ff.Command{
		Name:     "testcmd",
		Usage:    "testcmd [FLAGS] <SUBCOMMAND> ...",
		LongHelp: loremIpsum,
		FlagSet:  rootFlagSet,
	}

	fooFlagSet := ff.NewSet("foo").SetParent(rootFlagSet)
	fooFlagSet.IntVar(&vars.Alpha, 'a', "alpha", 10, "alpha integer")
	fooFlagSet.BoolVar(&vars.Beta, 'b', "beta", false, "beta boolean")
	fooCommand := &ff.Command{
		Name:      "foo",
		Usage:     "foo [FLAGS] <SUBCOMMAND> ...",
		ShortHelp: "the foo subcommand",
		FlagSet:   fooFlagSet,
		Exec:      func(_ context.Context, args []string) error { t.Logf("foo %+v %#v", vars, args); return nil },
	}
	rootCommand.Subcommands = append(rootCommand.Subcommands, fooCommand)

	barFlagSet := ff.NewSet("bar").SetParent(fooFlagSet)
	barFlagSet.DurationVar(&vars.Delta, 'd', "delta", 3*time.Second, "delta `δ` duration")
	barFlagSet.Float64Var(&vars.Epsilon, 'e', "epsilon", 3.21, "epsilon float")
	barCommand := &ff.Command{
		Name:      "bar",
		Usage:     "bar [FLAGS] ...",
		ShortHelp: "the bar subcommand",
		FlagSet:   barFlagSet,
		Exec:      func(_ context.Context, args []string) error { t.Logf("bar %+v %#v", vars, args); return nil },
	}
	fooCommand.Subcommands = append(fooCommand.Subcommands, barCommand)

	return rootCommand, &vars
}

type testCommandVars struct {
	Verbose    bool
	ConfigFile string
	Alpha      int
	Beta       bool
	Delta      time.Duration
	Epsilon    float64
}

var loremIpsum = strings.TrimSpace(`
Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nam diam eros,
vestibulum at pulvinar vulputate, vehicula id lacus. Class aptent taciti
sociosqu ad litora torquent per conubia nostra, per inceptos himenaeos.
Mauris venenatis felis orci, ac consectetur mi molestie ac. Integer pharetra
pharetra odio. Maecenas metus eros, viverra eget efficitur ut, feugiat in
tortor. Quisque elit nibh, rhoncus in posuere et, bibendum non turpis.
Maecenas eget dui malesuada, pretium tellus quis, bibendum felis. Duis erat
enim, faucibus id auctor ac, ornare sed metus.
`)

func compareTestCommandVars(t *testing.T, want, have testCommandVars) {
	t.Helper()
	var (
		structType = reflect.TypeOf(testCommandVars{})
		wantStruct = reflect.ValueOf(want)
		haveStruct = reflect.ValueOf(have)
	)
	for _, f := range reflect.VisibleFields(structType) {
		var (
			wantValue = wantStruct.FieldByIndex(f.Index)
			haveValue = haveStruct.FieldByIndex(f.Index)
		)
		if !wantValue.Equal(haveValue) {
			t.Errorf("%s: want %#v, have %#v", f.Name, wantValue, haveValue)
		}
	}
}
