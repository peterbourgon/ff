package ffcli_test

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/peterbourgon/ff/v3/fftest"
)

func TestCommandRun(t *testing.T) {
	t.Parallel()

	for _, testcase := range []struct {
		name     string
		args     []string
		rootvars fftest.Vars
		rootran  bool
		rootargs []string
		foovars  fftest.Vars
		fooran   bool
		fooargs  []string
		barvars  fftest.Vars
		barran   bool
		barargs  []string
	}{
		{
			name:    "root",
			rootran: true,
		},
		{
			name:     "root flags",
			args:     []string{"-s", "123", "-b"},
			rootvars: fftest.Vars{S: "123", B: true},
			rootran:  true,
		},
		{
			name:     "root args",
			args:     []string{"hello"},
			rootran:  true,
			rootargs: []string{"hello"},
		},
		{
			name:     "root flags args",
			args:     []string{"-i=123", "hello world"},
			rootvars: fftest.Vars{I: 123},
			rootran:  true,
			rootargs: []string{"hello world"},
		},
		{
			name:     "root flags -- args",
			args:     []string{"-f", "1.23", "--", "hello", "world"},
			rootvars: fftest.Vars{F: 1.23},
			rootran:  true,
			rootargs: []string{"hello", "world"},
		},
		{
			name:   "root foo",
			args:   []string{"foo"},
			fooran: true,
		},
		{
			name:     "root flags foo",
			args:     []string{"-s", "OK", "-d", "10m", "foo"},
			rootvars: fftest.Vars{S: "OK", D: 10 * time.Minute},
			fooran:   true,
		},
		{
			name:     "root flags foo flags",
			args:     []string{"-s", "OK", "-d", "10m", "foo", "-s", "Yup"},
			rootvars: fftest.Vars{S: "OK", D: 10 * time.Minute},
			foovars:  fftest.Vars{S: "Yup"},
			fooran:   true,
		},
		{
			name:     "root flags foo flags args",
			args:     []string{"-f=0.99", "foo", "-f", "1.01", "verb", "noun", "adjective adjective"},
			rootvars: fftest.Vars{F: 0.99},
			foovars:  fftest.Vars{F: 1.01},
			fooran:   true,
			fooargs:  []string{"verb", "noun", "adjective adjective"},
		},
		{
			name:     "root flags foo args",
			args:     []string{"-f=0.99", "foo", "abc", "def", "ghi"},
			rootvars: fftest.Vars{F: 0.99},
			fooran:   true,
			fooargs:  []string{"abc", "def", "ghi"},
		},
		{
			name:    "root bar -- args",
			args:    []string{"bar", "--", "argument", "list"},
			barran:  true,
			barargs: []string{"argument", "list"},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			foofs, foovars := fftest.Pair()
			var fooargs []string
			var fooran bool
			foo := &ffcli.Command{
				Name:    "foo",
				FlagSet: foofs,
				Exec:    func(_ context.Context, args []string) error { fooran, fooargs = true, args; return nil },
			}

			barfs, barvars := fftest.Pair()
			var barargs []string
			var barran bool
			bar := &ffcli.Command{
				Name:    "bar",
				FlagSet: barfs,
				Exec:    func(_ context.Context, args []string) error { barran, barargs = true, args; return nil },
			}

			rootfs, rootvars := fftest.Pair()
			var rootargs []string
			var rootran bool
			root := &ffcli.Command{
				FlagSet:     rootfs,
				Subcommands: []*ffcli.Command{foo, bar},
				Exec:        func(_ context.Context, args []string) error { rootran, rootargs = true, args; return nil },
			}

			err := root.ParseAndRun(context.Background(), testcase.args)
			assertNoError(t, err)
			fftest.Compare(t, &testcase.rootvars, rootvars)
			assertBool(t, testcase.rootran, rootran)
			assertStringSlice(t, testcase.rootargs, rootargs)
			fftest.Compare(t, &testcase.foovars, foovars)
			assertBool(t, testcase.fooran, fooran)
			assertStringSlice(t, testcase.fooargs, fooargs)
			fftest.Compare(t, &testcase.barvars, barvars)
			assertBool(t, testcase.barran, barran)
			assertStringSlice(t, testcase.barargs, barargs)
		})
	}
}

func TestHelpUsage(t *testing.T) {
	t.Parallel()

	for _, testcase := range []struct {
		name      string
		usageFunc func(*ffcli.Command) string
		exec      func(context.Context, []string) error
		args      []string
		output    string
	}{
		{
			name:   "nil",
			args:   []string{"-h"},
			output: defaultUsageFuncOutput,
		},
		{
			name:      "DefaultUsageFunc",
			usageFunc: ffcli.DefaultUsageFunc,
			args:      []string{"-h"},
			output:    defaultUsageFuncOutput,
		},
		{
			name:      "custom usage",
			usageFunc: func(*ffcli.Command) string { return "🍰" },
			args:      []string{"-h"},
			output:    "🍰\n",
		},
		{
			name:      "ErrHelp",
			usageFunc: func(*ffcli.Command) string { return "👹" },
			exec:      func(context.Context, []string) error { return flag.ErrHelp },
			output:    "👹\n",
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			fs, _ := fftest.Pair()
			var buf bytes.Buffer
			fs.SetOutput(&buf)

			command := &ffcli.Command{
				Name:       "TestHelpUsage",
				ShortUsage: "TestHelpUsage [flags] <args>",
				ShortHelp:  "Some short help.",
				LongHelp:   "Some long help.",
				FlagSet:    fs,
				UsageFunc:  testcase.usageFunc,
				Exec:       testcase.exec,
			}

			err := command.ParseAndRun(context.Background(), testcase.args)
			assertErrorIs(t, flag.ErrHelp, err)
			assertMultilineString(t, testcase.output, buf.String())
		})
	}
}

func TestNestedOutput(t *testing.T) {
	t.Parallel()

	for _, testcase := range []struct {
		name       string
		args       []string
		wantErr    error
		wantOutput string
	}{
		{
			name:       "root without args",
			args:       []string{},
			wantErr:    flag.ErrHelp,
			wantOutput: "root usage func\n",
		},
		{
			name:       "root with args",
			args:       []string{"abc", "def ghi"},
			wantErr:    flag.ErrHelp,
			wantOutput: "root usage func\n",
		},
		{
			name:       "root help",
			args:       []string{"-h"},
			wantErr:    flag.ErrHelp,
			wantOutput: "root usage func\n",
		},
		{
			name:       "foo without args",
			args:       []string{"foo"},
			wantOutput: "foo: ''\n",
		},
		{
			name:       "foo with args",
			args:       []string{"foo", "alpha", "beta"},
			wantOutput: "foo: 'alpha beta'\n",
		},
		{
			name:       "foo help",
			args:       []string{"foo", "-h"},
			wantErr:    flag.ErrHelp,
			wantOutput: "foo usage func\n", // only one instance of usage string
		},
		{
			name:       "foo bar without args",
			args:       []string{"foo", "bar"},
			wantErr:    flag.ErrHelp,
			wantOutput: "bar usage func\n",
		},
		{
			name:       "foo bar with args",
			args:       []string{"foo", "bar", "--", "baz quux"},
			wantErr:    flag.ErrHelp,
			wantOutput: "bar usage func\n",
		},
		{
			name:       "foo bar help",
			args:       []string{"foo", "bar", "--help"},
			wantErr:    flag.ErrHelp,
			wantOutput: "bar usage func\n",
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			var (
				rootfs = flag.NewFlagSet("root", flag.ContinueOnError)
				foofs  = flag.NewFlagSet("foo", flag.ContinueOnError)
				barfs  = flag.NewFlagSet("bar", flag.ContinueOnError)
				buf    bytes.Buffer
			)
			rootfs.SetOutput(&buf)
			foofs.SetOutput(&buf)
			barfs.SetOutput(&buf)

			barExec := func(_ context.Context, args []string) error {
				return flag.ErrHelp
			}

			bar := &ffcli.Command{
				Name:      "bar",
				FlagSet:   barfs,
				UsageFunc: func(*ffcli.Command) string { return "bar usage func" },
				Exec:      barExec,
			}

			fooExec := func(_ context.Context, args []string) error {
				fmt.Fprintf(&buf, "foo: '%s'\n", strings.Join(args, " "))
				return nil
			}

			foo := &ffcli.Command{
				Name:        "foo",
				FlagSet:     foofs,
				UsageFunc:   func(*ffcli.Command) string { return "foo usage func" },
				Subcommands: []*ffcli.Command{bar},
				Exec:        fooExec,
			}

			rootExec := func(_ context.Context, args []string) error {
				return flag.ErrHelp
			}

			root := &ffcli.Command{
				FlagSet:     rootfs,
				UsageFunc:   func(*ffcli.Command) string { return "root usage func" },
				Subcommands: []*ffcli.Command{foo},
				Exec:        rootExec,
			}

			err := root.ParseAndRun(context.Background(), testcase.args)
			if want, have := testcase.wantErr, err; !errors.Is(have, want) {
				t.Errorf("error: want %v, have %v", want, have)
			}
			if want, have := testcase.wantOutput, buf.String(); want != have {
				t.Errorf("output: want %q, have %q", want, have)
			}
		})
	}
}

func TestIssue57(t *testing.T) {
	t.Parallel()

	for _, testcase := range []struct {
		args        []string
		parseErrAs  any
		parseErrIs  error
		parseErrStr string
		runErrAs    any
		runErrIs    error
		runErrStr   string
	}{
		{
			args:       []string{},
			parseErrAs: &ffcli.NoExecError{},
			runErrAs:   &ffcli.NoExecError{},
		},
		{
			args:       []string{"-h"},
			parseErrIs: flag.ErrHelp,
			runErrIs:   ffcli.ErrUnparsed,
		},
		{
			args:       []string{"bar"},
			parseErrAs: &ffcli.NoExecError{},
			runErrAs:   &ffcli.NoExecError{},
		},
		{
			args:       []string{"bar", "-h"},
			parseErrAs: flag.ErrHelp,
			runErrAs:   ffcli.ErrUnparsed,
		},
		{
			args:        []string{"bar", "-undefined"},
			parseErrStr: "error parsing commandline arguments: flag provided but not defined: -undefined",
			runErrIs:    ffcli.ErrUnparsed,
		},
		{
			args: []string{"bar", "baz"},
		},
		{
			args:       []string{"bar", "baz", "-h"},
			parseErrIs: flag.ErrHelp,
			runErrIs:   ffcli.ErrUnparsed,
		},
		{
			args:        []string{"bar", "baz", "-also.undefined"},
			parseErrStr: "error parsing commandline arguments: flag provided but not defined: -also.undefined",
			runErrIs:    ffcli.ErrUnparsed,
		},
	} {
		t.Run(strings.Join(append([]string{"foo"}, testcase.args...), " "), func(t *testing.T) {
			fs := flag.NewFlagSet("·", flag.ContinueOnError)
			fs.SetOutput(io.Discard)

			var (
				baz = &ffcli.Command{Name: "baz", FlagSet: fs, Exec: func(_ context.Context, args []string) error { return nil }}
				bar = &ffcli.Command{Name: "bar", FlagSet: fs, Subcommands: []*ffcli.Command{baz}}
				foo = &ffcli.Command{Name: "foo", FlagSet: fs, Subcommands: []*ffcli.Command{bar}}
			)

			var (
				parseErr = foo.Parse(testcase.args)
				runErr   = foo.Run(context.Background())
			)

			if testcase.parseErrAs != nil {
				if want, have := &testcase.parseErrAs, parseErr; !errors.As(have, want) {
					t.Errorf("Parse: want %v, have %v", want, have)
				}
			}

			if testcase.parseErrIs != nil {
				if want, have := testcase.parseErrIs, parseErr; !errors.Is(have, want) {
					t.Errorf("Parse: want %v, have %v", want, have)
				}
			}

			if testcase.parseErrStr != "" {
				if want, have := testcase.parseErrStr, parseErr.Error(); want != have {
					t.Errorf("Parse: want %q, have %q", want, have)
				}
			}

			if testcase.runErrAs != nil {
				if want, have := &testcase.runErrAs, runErr; !errors.As(have, want) {
					t.Errorf("Run: want %v, have %v", want, have)
				}
			}

			if testcase.runErrIs != nil {
				if want, have := testcase.runErrIs, runErr; !errors.Is(have, want) {
					t.Errorf("Run: want %v, have %v", want, have)
				}
			}

			if testcase.runErrStr != "" {
				if want, have := testcase.runErrStr, runErr.Error(); want != have {
					t.Errorf("Run: want %q, have %q", want, have)
				}
			}

			var (
				noParseErr = testcase.parseErrAs == nil && testcase.parseErrIs == nil && testcase.parseErrStr == ""
				noRunErr   = testcase.runErrAs == nil && testcase.runErrIs == nil && testcase.runErrStr == ""
			)
			if noParseErr && noRunErr {
				if parseErr != nil {
					t.Errorf("Parse: unexpected error: %v", parseErr)
				}
				if runErr != nil {
					t.Errorf("Run: unexpected error: %v", runErr)
				}
			}
		})
	}
}

func TestDefaultUsageFuncFlagHelp(t *testing.T) {
	t.Parallel()

	for _, testcase := range []struct {
		name string // name of test case
		def  string // default value, if any
		help string // help text for flag
		want string // expected usage text
	}{
		{
			name: "plain text",
			help: "does stuff",
			want: "-x string  does stuff",
		},
		{
			name: "placeholder",
			help: "reads from `file` instead of stdout",
			want: "-x file  reads from file instead of stdout",
		},
		{
			name: "default",
			def:  "www",
			help: "path to output directory",
			want: "-x www  path to output directory",
		},
		{
			name: "default with placeholder",
			def:  "www",
			help: "path to output `directory`",
			want: "-x www  path to output directory",
		},
	} {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			fset := flag.NewFlagSet(t.Name(), flag.ContinueOnError)
			fset.String("x", testcase.def, testcase.help)

			usage := ffcli.DefaultUsageFunc(&ffcli.Command{
				FlagSet: fset,
			})

			// Discard everything before the FLAGS section.
			_, flagUsage, ok := strings.Cut(usage, "\nFLAGS\n")
			if !ok {
				t.Fatalf("FLAGS section not found in:\n%s", usage)
			}

			assertMultilineString(t,
				strings.TrimSpace(testcase.want),
				strings.TrimSpace(flagUsage))
		})
	}
}

func ExampleCommand_Parse_then_Run() {
	// Assume our CLI will use some client that requires a token.
	type FooClient struct {
		token string
	}

	// That client would have a constructor.
	NewFooClient := func(token string) (*FooClient, error) {
		if token == "" {
			return nil, fmt.Errorf("token required")
		}
		return &FooClient{token: token}, nil
	}

	// We define the token in the root command's FlagSet.
	var (
		rootFlagSet = flag.NewFlagSet("mycommand", flag.ExitOnError)
		token       = rootFlagSet.String("token", "", "API token")
	)

	// Create a placeholder client, initially nil.
	var client *FooClient

	// Commands can reference and use it, because by the time their Exec
	// function is invoked, the client will be constructed.
	foo := &ffcli.Command{
		Name: "foo",
		Exec: func(context.Context, []string) error {
			fmt.Printf("subcommand foo can use the client: %v", client)
			return nil
		},
	}

	root := &ffcli.Command{
		FlagSet:     rootFlagSet,
		Subcommands: []*ffcli.Command{foo},
	}

	// Call Parse first, to populate flags and select a terminal command.
	if err := root.Parse([]string{"-token", "SECRETKEY", "foo"}); err != nil {
		log.Fatalf("Parse failure: %v", err)
	}

	// After a successful Parse, we can construct a FooClient with the token.
	var err error
	client, err = NewFooClient(*token)
	if err != nil {
		log.Fatalf("error constructing FooClient: %v", err)
	}

	// Then call Run, which will select the foo subcommand and invoke it.
	if err := root.Run(context.Background()); err != nil {
		log.Fatalf("Run failure: %v", err)
	}

	// Output:
	// subcommand foo can use the client: &{SECRETKEY}
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func assertErrorIs(t *testing.T, want, have error) {
	t.Helper()
	if !errors.Is(have, want) {
		t.Fatalf("want %v, have %v", want, have)
	}
}

func assertMultilineString(t *testing.T, want, have string) {
	t.Helper()
	if want != have {
		t.Fatalf("\nwant:\n%s\n\nhave:\n%s\n", want, have)
	}
}

func assertBool(t *testing.T, want, have bool) {
	t.Helper()
	if want != have {
		t.Fatalf("want %v, have %v", want, have)
	}
}

func assertStringSlice(t *testing.T, want, have []string) {
	t.Helper()
	if len(want) == 0 && len(have) == 0 {
		return // consider []string{} and []string(nil) equivalent
	}
	if !reflect.DeepEqual(want, have) {
		t.Fatalf("want %#+v, have %#+v", want, have)
	}
}

var defaultUsageFuncOutput = strings.TrimSpace(`
DESCRIPTION
  Some short help.

USAGE
  TestHelpUsage [flags] <args>

Some long help.

FLAGS
  -b=false   bool
  -d 0s      time.Duration
  -f 0       float64
  -i 0       int
  -s string  string
  -x ...     collection of strings (repeatable)
`) + "\n\n"
