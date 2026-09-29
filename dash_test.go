package ff_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/fftest"
)

// A single "-" is a non-flag argument: parsing stops and GetArgs retains it,
// along with the remaining arguments, for both flag set constructors.
func TestFlagSet_SingleDashTerminator(t *testing.T) {
	t.Parallel()

	constructors := []fftest.Constructor{fftest.CoreConstructor, fftest.StdConstructor}

	testcases := fftest.TestCases{
		{
			Name:         "single dash alone is a positional arg",
			Constructors: constructors,
			Args:         []string{"-"},
			Want:         fftest.Vars{Args: []string{"-"}},
		},
		{
			Name:         "single dash terminates parsing of following flags",
			Constructors: constructors,
			Args:         []string{"-", "-b"},
			Want:         fftest.Vars{B: false, Args: []string{"-", "-b"}},
		},
		{
			Name:         "flags before a single dash still parse, dash remains",
			Constructors: constructors,
			Args:         []string{"-b", "-"},
			Want:         fftest.Vars{B: true, Args: []string{"-"}},
		},
		{
			Name:         "single dash stops a following flag and its value",
			Constructors: constructors,
			Args:         []string{"-", "-s", "x"},
			Want:         fftest.Vars{S: "", Args: []string{"-", "-s", "x"}},
		},
		{
			Name:         "single dash after a value flag is consumed as its value",
			Constructors: constructors,
			Args:         []string{"-s", "-"},
			Want:         fftest.Vars{S: "-", Args: []string{}},
		},
		{
			Name:         "single dash as a value leaves following positional args",
			Constructors: constructors,
			Args:         []string{"-s", "-", "input", "-b"},
			Want:         fftest.Vars{S: "-", Args: []string{"input", "-b"}},
		},
		{
			Name:         "single dash as a value allows following flags to parse",
			Constructors: constructors,
			Args:         []string{"-s", "-", "-b"},
			Want:         fftest.Vars{S: "-", B: true, Args: []string{}},
		},
		{
			Name:         "single dash suppresses an unknown-flag error after it",
			Constructors: constructors,
			Args:         []string{"-", "--nope"},
			Want:         fftest.Vars{Args: []string{"-", "--nope"}},
		},
		{
			Name:         "unknown flags before a single dash still return an error",
			Constructors: constructors,
			Args:         []string{"--nope", "-"},
			Want:         fftest.Vars{WantParseErrorIs: ff.ErrUnknownFlag},
		},
		{
			Name:         "double dash terminator is excluded from positional args",
			Constructors: constructors,
			Args:         []string{"--", "-", "-b"},
			Want:         fftest.Vars{Args: []string{"-", "-b"}},
		},
		{
			Name:         "ordinary positional args still stop flag parsing",
			Constructors: constructors,
			Args:         []string{"input", "-", "-b"},
			Want:         fftest.Vars{Args: []string{"input", "-", "-b"}},
		},
		{
			Name:         "empty positional args still stop flag parsing",
			Constructors: constructors,
			Args:         []string{"", "-", "-b"},
			Want:         fftest.Vars{Args: []string{"", "-", "-b"}},
		},
	}

	testcases.Run(t)
}

func TestCommand_SingleDashArgument(t *testing.T) {
	t.Parallel()

	for _, constructor := range []fftest.Constructor{fftest.CoreConstructor, fftest.StdConstructor} {
		t.Run(constructor.Name, func(t *testing.T) {
			fs, vars := constructor.Make(fftest.Vars{})
			subfs, _ := constructor.Make(fftest.Vars{})
			var execArgs []string
			exec := func(_ context.Context, args []string) error {
				execArgs = args
				return nil
			}
			cmd := &ff.Command{
				Name:  "root",
				Flags: fs,
				Subcommands: []*ff.Command{
					{Name: "read", Flags: subfs, Exec: exec},
				},
				Exec: exec,
			}
			args := []string{"-", "read", "-b"}
			if err := cmd.ParseAndRun(context.Background(), args); err != nil {
				t.Fatalf("parse and run: %v", err)
			}
			if cmd.GetSelected() != cmd {
				t.Errorf("selected command: want root (%p), have %p", cmd, cmd.GetSelected())
			}
			if !reflect.DeepEqual(execArgs, args) {
				t.Errorf("exec args: want %v, have %v", args, execArgs)
			}
			if vars.B {
				t.Error("flag after a positional argument was parsed")
			}
		})
	}
}
