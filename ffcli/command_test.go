package ffcli_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/peterbourgon/ff"
	"github.com/peterbourgon/ff/ffcli"
	"github.com/peterbourgon/ff/fftest"
)

func TestCommandParseAndRun(t *testing.T) {
	for _, testcase := range []struct {
		name     string
		args     []string
		options  []ff.Option
		rootvars fftest.Vars
		rootargs []string
		foovars  fftest.Vars
		fooargs  []string
		barvars  fftest.Vars
		barargs  []string
	}{
		/*{ // should work but doesn't!! :[
			name:     "root",
			rootargs: []string{},
		}*/
		{
			name:     "root flags",
			args:     []string{"-s", "123", "-b"},
			rootvars: fftest.Vars{S: "123", B: true},
			rootargs: []string{},
		},
		{
			name:     "root args",
			args:     []string{"hello"},
			rootargs: []string{"hello"},
		},
		{
			name:     "root flags args",
			args:     []string{"-i=123", "hello world"},
			rootvars: fftest.Vars{I: 123},
			rootargs: []string{"hello world"},
		},
		{
			name:     "root flags -- args",
			args:     []string{"-f", "1.23", "--", "hello", "world"},
			rootvars: fftest.Vars{F: 1.23},
			rootargs: []string{"hello", "world"},
		},
		{
			name:    "root foo",
			args:    []string{"foo"},
			fooargs: []string{},
		},
		{
			name:     "root flags foo",
			args:     []string{"-s", "OK", "-d", "10m", "foo"},
			rootvars: fftest.Vars{S: "OK", D: 10 * time.Minute},
			fooargs:  []string{},
		},
		{
			name:     "root flags foo flags",
			args:     []string{"-s", "OK", "-d", "10m", "foo", "-s", "Yup"},
			rootvars: fftest.Vars{S: "OK", D: 10 * time.Minute},
			foovars:  fftest.Vars{S: "Yup"},
			fooargs:  []string{},
		},
		{
			name:     "root flags foo flags args",
			args:     []string{"-f=0.99", "foo", "-f", "1.01", "verb", "noun", "adjective adjective"},
			rootvars: fftest.Vars{F: 0.99},
			foovars:  fftest.Vars{F: 1.01},
			fooargs:  []string{"verb", "noun", "adjective adjective"},
		},
		{
			name:     "root flags foo args",
			args:     []string{"-f=0.99", "foo", "abc", "def", "ghi"},
			rootvars: fftest.Vars{F: 0.99},
			fooargs:  []string{"abc", "def", "ghi"},
		},
		{
			name:    "root bar -- args",
			args:    []string{"bar", "--", "argument", "list"},
			barargs: []string{"argument", "list"},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			foofs, foovars := fftest.Pair()
			var fooargs []string
			foo := &ffcli.Command{
				Name:    "foo",
				FlagSet: foofs,
				Main:    func(args []string) error { fooargs = args; return nil },
			}

			barfs, barvars := fftest.Pair()
			var barargs []string
			bar := &ffcli.Command{
				Name:    "bar",
				FlagSet: barfs,
				Main:    func(args []string) error { barargs = args; return nil },
			}

			rootfs, rootvars := fftest.Pair()
			var rootargs []string
			root := &ffcli.Command{
				FlagSet:     rootfs,
				Subcommands: []*ffcli.Command{foo, bar},
				Main:        func(args []string) error { rootargs = args; return nil },
			}

			err := root.ParseAndRun(testcase.args, testcase.options...)

			assertNoError(t, err)
			assertNoError(t, fftest.Compare(&testcase.rootvars, rootvars))
			assertStringSlice(t, testcase.rootargs, rootargs)
			assertNoError(t, fftest.Compare(&testcase.foovars, foovars))
			assertStringSlice(t, testcase.fooargs, fooargs)
			assertNoError(t, fftest.Compare(&testcase.barvars, barvars))
			assertStringSlice(t, testcase.barargs, barargs)
		})
	}
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
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
	if !reflect.DeepEqual(want, have) {
		t.Fatalf("want %v, have %v", want, have)
	}
}
