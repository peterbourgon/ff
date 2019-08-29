package ffcli_test

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/peterbourgon/ff/ffcli"
	"github.com/peterbourgon/ff/fftest"
)

func TestCommandRun(t *testing.T) {
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

			err := root.Run(context.Background(), testcase.args)
			assertNoError(t, err)
			assertNoError(t, fftest.Compare(&testcase.rootvars, rootvars))
			assertBool(t, testcase.rootran, rootran)
			assertStringSlice(t, testcase.rootargs, rootargs)
			assertNoError(t, fftest.Compare(&testcase.foovars, foovars))
			assertBool(t, testcase.fooran, fooran)
			assertStringSlice(t, testcase.fooargs, fooargs)
			assertNoError(t, fftest.Compare(&testcase.barvars, barvars))
			assertBool(t, testcase.barran, barran)
			assertStringSlice(t, testcase.barargs, barargs)
		})
	}
}

func TestHelpUsage(t *testing.T) {
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
				Name:      "TestUsageFunc",
				Usage:     "TestUsageFunc [flags] <args>",
				ShortHelp: "Some short help.",
				LongHelp:  "Some long help.",
				FlagSet:   fs,
				UsageFunc: testcase.usageFunc,
				Exec:      testcase.exec,
			}

			err := command.Run(context.Background(), testcase.args)
			assertError(t, flag.ErrHelp, err)
			assertString(t, testcase.output, buf.String())
		})
	}
}

func ExampleCommand_Postparse() {
	// Assume our CLI will use some client that requires a token.
	type FooClient struct {
		token string
	}

	// It'll have a constructor.
	NewFooClient := func(token string) *FooClient {
		return &FooClient{token: token}
	}

	// We would define that token in the root command's FlagSet.
	var (
		rootFlagSet = flag.NewFlagSet("mycommand", flag.ExitOnError)
		token       = rootFlagSet.String("token", "", "API token")
	)

	// Create a placeholder client.
	var client *FooClient

	// That client will be usable by subcommands...
	foo := &ffcli.Command{
		Name: "foo",
		Exec: func(context.Context, []string) error {
			fmt.Printf("subcommand foo can use the client: %v", client)
			return nil
		},
	}

	// ...as long as we set it up in the root command's Postparse.
	root := &ffcli.Command{
		FlagSet: rootFlagSet,
		Postparse: func(context.Context) error {
			client = NewFooClient(*token)
			return nil
		},
		Subcommands: []*ffcli.Command{foo},
	}

	root.Run(context.Background(), []string{"-token", "SECRETKEY", "foo"})

	// Output:
	// subcommand foo can use the client: &{SECRETKEY}
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func assertError(t *testing.T, want, have error) {
	t.Helper()
	if want != have {
		t.Fatalf("want %v, have %v", want, have)
	}
}

func assertString(t *testing.T, want, have string) {
	t.Helper()
	if want != have {
		t.Fatalf("want %q, have %q", want, have)
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
USAGE
  TestUsageFunc [flags] <args>

Some long help.

FLAGS
  -b false  bool
  -d 0s     time.Duration
  -f 0      float64
  -i 0      int
  -s ...    string
  -x ...    collection of strings (repeatable)
`) + "\n"
