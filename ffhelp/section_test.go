package ffhelp_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffhelp"
	"github.com/peterbourgon/ff/v4/fftest"
)

func TestSection_Flags(t *testing.T) {
	t.Parallel()

	t.Run("default", func(t *testing.T) {
		fs, _ := fftest.CoreConstructor.Make(fftest.Vars{A: true})
		want := strings.TrimSpace(coreFlagsDefaultFlagsSection)
		have := strings.TrimSpace(ffhelp.NewFlagsSection(fs).String())
		if want != have {
			t.Errorf("\n%s", fftest.DiffString(want, have))
		}
	})
}

var coreFlagsDefaultFlagsSection = strings.TrimSpace(`
FLAGS
  -s, --str STRING     string
  -i, --int INT        int (default: 0)
  -f, --flt FLOAT64    float64 (default: 0)
  -a, --aflag BOOL     bool a (default: true)
  -b, --bflag          bool b (default: false)
  -c, --cflag          bool c (default: false)
  -d, --dur DURATION   time.Duration (default: 0s)
  -x, --xxx STR        collection of strings (repeatable)
`)

//
//
//

func TestSection_StdFlags(t *testing.T) {
	t.Parallel()

	fs, _ := fftest.StdConstructor.Make(fftest.Vars{A: true})
	want := strings.TrimSpace(stdFlagsSectionHelp)
	have := strings.TrimSpace(ffhelp.Flags(fs).String())
	if want != have {
		t.Errorf("\n%s", fftest.DiffString(want, have))
	}
}

var stdFlagsSectionHelp = `
fftest

FLAGS
  -a BOOL       bool a (default: true)
  -b            bool b (default: false)
  -c            bool c (default: false)
  -d DURATION   time.Duration (default: 0s)
  -f FLOAT64    float64 (default: 0)
  -i INT        int (default: 0)
  -s STRING     string
  -x STRING     collection of strings (repeatable)
`

//
//
//

func TestSections_Command(t *testing.T) {
	t.Parallel()

	t.Run("unparsed", func(t *testing.T) {
		testcmd := makeTestCommand(t)
		want := strings.TrimSpace(testCommandRootHelp)
		have := strings.TrimSpace(ffhelp.Command(testcmd).String())
		if want != have {
			t.Errorf("\n%s", fftest.DiffString(want, have))
		}
	})

	for _, test := range []struct {
		name string
		args []string
		want string
	}{
		{
			name: "no args",
			args: []string{},
			want: testCommandRootHelp,
		},
		{
			name: "-h",
			args: []string{"-h"},
			want: testCommandRootHelp,
		},
		{
			name: "--help",
			args: []string{"--help"},
			want: testCommandRootHelp,
		},
		{
			name: "-v foo",
			args: []string{"-v", "foo"},
			want: testCommandFooHelp,
		},
		{
			name: "-v foo bar --alpha=9",
			args: []string{"-v", "foo", "bar", "--alpha=9"},
			want: testCommandBarHelp,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			testcmd := makeTestCommand(t)
			err := testcmd.ParseAndRun(context.Background(), test.args)
			switch {
			case err == nil, errors.Is(err, ff.ErrHelp), errors.Is(err, ff.ErrNoExec):
				// ok
			default:
				t.Fatal(err)
			}

			want := strings.TrimSpace(test.want)
			have := strings.TrimSpace(ffhelp.Command(testcmd).String())
			if want != have {
				t.Errorf("\n%s", fftest.DiffString(want, have))
			}
		})
	}
}

var testCommandRootHelp = `
testcmd

USAGE
  testcmd [FLAGS] <SUBCOMMAND> ...

Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nam diam eros,
vestibulum at pulvinar vulputate, vehicula id lacus. Class aptent taciti
sociosqu ad litora torquent per conubia nostra, per inceptos himenaeos.
Mauris venenatis felis orci, ac consectetur mi molestie ac. Integer pharetra
pharetra odio. Maecenas metus eros, viverra eget efficitur ut, feugiat in
tortor. Quisque elit nibh, rhoncus in posuere et, bibendum non turpis.
Maecenas eget dui malesuada, pretium tellus quis, bibendum felis. Duis erat
enim, faucibus id auctor ac, ornare sed metus.

SUBCOMMANDS
  foo   the foo subcommand

FLAGS
  -v, --verbose              verbose logging (default: false)
      --config-file STRING   config file
`

var testCommandFooHelp = `
foo -- the foo subcommand

USAGE
  foo [FLAGS] <SUBCOMMAND> ...

SUBCOMMANDS
  bar   the bar subcommand

FLAGS
  -a, --alpha INT            alpha integer (default: 10)
  -b, --beta                 beta boolean (default: false)

FLAGS (root)
  -v, --verbose              verbose logging (default: false)
      --config-file STRING   config file
`

var testCommandBarHelp = strings.ReplaceAll(`
bar -- the bar subcommand

USAGE
  bar [FLAGS] ...

FLAGS
  -d, --delta δ              delta #δ# duration (default: 3s)
  -e, --epsilon FLOAT64      epsilon float (default: 3.21)

FLAGS (foo)
  -a, --alpha INT            alpha integer (default: 10)
  -b, --beta                 beta boolean (default: false)

FLAGS (root)
  -v, --verbose              verbose logging (default: false)
      --config-file STRING   config file
`, "#", "`")

//
//
//

func makeTestCommand(t *testing.T) *ff.Command {
	t.Helper()

	rootFlags := ff.NewFlags("root")
	rootFlags.Bool('v', "verbose", false, "verbose logging")
	rootFlags.String(0, "config-file", "", "config file")
	rootCommand := &ff.Command{
		Name:     "testcmd",
		Usage:    "testcmd [FLAGS] <SUBCOMMAND> ...",
		LongHelp: strings.TrimSpace(loremIpsum),
		Flags:    rootFlags,
	}

	fooFlags := ff.NewFlags("foo").SetParent(rootFlags)
	fooFlags.Int('a', "alpha", 10, "alpha integer")
	fooFlags.Bool('b', "beta", false, "beta boolean")
	fooCommand := &ff.Command{
		Name:      "foo",
		Usage:     "foo [FLAGS] <SUBCOMMAND> ...",
		ShortHelp: "the foo subcommand",
		Flags:     fooFlags,
	}
	rootCommand.Subcommands = append(rootCommand.Subcommands, fooCommand)

	barFlags := ff.NewFlags("bar").SetParent(fooFlags)
	barFlags.Duration('d', "delta", 3*time.Second, "delta `δ` duration")
	barFlags.Float64('e', "epsilon", 3.21, "epsilon float")
	barCommand := &ff.Command{
		Name:      "bar",
		Usage:     "bar [FLAGS] ...",
		ShortHelp: "the bar subcommand",
		Flags:     barFlags,
	}
	fooCommand.Subcommands = append(fooCommand.Subcommands, barCommand)

	return rootCommand
}

var loremIpsum = `
Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nam diam eros,
vestibulum at pulvinar vulputate, vehicula id lacus. Class aptent taciti
sociosqu ad litora torquent per conubia nostra, per inceptos himenaeos.
Mauris venenatis felis orci, ac consectetur mi molestie ac. Integer pharetra
pharetra odio. Maecenas metus eros, viverra eget efficitur ut, feugiat in
tortor. Quisque elit nibh, rhoncus in posuere et, bibendum non turpis.
Maecenas eget dui malesuada, pretium tellus quis, bibendum felis. Duis erat
enim, faucibus id auctor ac, ornare sed metus.
`
