package ff_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/fftest"
)

func TestDefaultFlagSetUsage(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		constr fftest.Constructor
		want   string
	}{
		{fftest.CoreConstructor, coreFlagSetDefaultUsage},
		{fftest.StdConstructor, stdFlagSetDefaultUsage},
	} {
		t.Run(test.constr.Name, func(t *testing.T) {
			fs, _ := test.constr.Make(fftest.Vars{A: true})
			want := strings.TrimSpace(test.want)
			have := strings.TrimSpace(ff.DefaultFlagSetUsage(fs))
			if want != have {
				t.Errorf("\n%s", fftest.DiffString(want, have))
			}
		})
	}
}

var coreFlagSetDefaultUsage = strings.TrimSpace(`
COMMAND
  fftest

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

var stdFlagSetDefaultUsage = strings.TrimSpace(`
COMMAND
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
`)

func TestDefaultCommandUsage(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name string
		args []string
		want string
	}{
		{
			name: "no args",
			args: []string{},
			want: testCommandRootUsage,
		},
		{
			name: "-h",
			args: []string{"-h"},
			want: testCommandRootUsage,
		},
		{
			name: "--help",
			args: []string{"--help"},
			want: testCommandRootUsage,
		},
		{
			name: "-v foo",
			args: []string{"-v", "foo"},
			want: testCommandFooUsage,
		},
		{
			name: "-v foo bar --alpha=9",
			args: []string{"-v", "foo", "bar", "--alpha=9"},
			want: testCommandBarUsage,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			testcmd, _ := makeTestCommand(t)

			err := testcmd.ParseAndRun(ctx, test.args)
			switch {
			case err == nil, errors.Is(err, ff.ErrHelp), errors.Is(err, ff.ErrNoExec):
				// ok
			default:
				t.Fatal(err)
			}

			want := strings.TrimSpace(test.want)
			have := strings.TrimSpace(ff.DefaultCommandUsage(testcmd))
			if want != have {
				t.Errorf("\n%s", fftest.DiffString(want, have))
			}
		})
	}
}

var testCommandRootUsage = strings.TrimSpace(`
COMMAND
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
`)

var testCommandFooUsage = strings.TrimSpace(`
COMMAND
  foo -- the foo subcommand

USAGE
  foo [FLAGS] <SUBCOMMAND> ...

SUBCOMMANDS
  bar   the bar subcommand

FLAGS
  -a, --alpha INT   alpha integer (default: 10)
  -b, --beta        beta boolean (default: false)

FLAGS (root)
  -v, --verbose              verbose logging (default: false)
      --config-file STRING   config file
`)

var testCommandBarUsage = strings.TrimSpace(strings.ReplaceAll(`
COMMAND
  bar -- the bar subcommand

USAGE
  bar [FLAGS] ...

FLAGS
  -d, --delta δ           delta #δ# duration (default: 3s)
  -e, --epsilon FLOAT64   epsilon float (default: 3.21)

FLAGS (foo)
  -a, --alpha INT   alpha integer (default: 10)
  -b, --beta        beta boolean (default: false)

FLAGS (root)
  -v, --verbose              verbose logging (default: false)
      --config-file STRING   config file
`, "#", "`"))
