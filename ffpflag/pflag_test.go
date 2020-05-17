package ffpflag_test

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/peterbourgon/ff/v3"
	"github.com/peterbourgon/ff/v3/ffpflag"
	"github.com/peterbourgon/ff/v3/fftoml"
	"github.com/peterbourgon/ff/v3/ffyaml"
	"github.com/spf13/pflag"
)

type Vars struct {
	S    string
	B    bool
	F    float32
	I    []int32
	Args []string

	ParseError           error
	WantParseErrorIs     error
	WantParseErrorString string
}

func Pair() (*pflag.FlagSet, *Vars) {
	fs := pflag.NewFlagSet("ffpflag_test", pflag.ContinueOnError)

	var v Vars
	fs.StringVarP(&v.S, "string", "s", "", "a string value")
	fs.BoolVarP(&v.B, "bool", "b", false, "a bool value")
	fs.Float32Var(&v.F, "f", 0., "a float32 value")
	fs.Int32SliceVarP(&v.I, "int32", "i", nil, "collection of int32 (repeatable)")

	return fs, &v
}

func Compare(want, have *Vars) error {
	// Normalize args.
	if want.Args == nil {
		want.Args = []string{}
	}
	if have.Args == nil {
		have.Args = []string{}
	}

	if want.WantParseErrorIs != nil || want.WantParseErrorString != "" {
		if want.WantParseErrorIs != nil && have.ParseError == nil {
			return fmt.Errorf("want error (%v), have none", want.WantParseErrorIs)
		}

		if want.WantParseErrorString != "" && have.ParseError == nil {
			return fmt.Errorf("want error (%q), have none", want.WantParseErrorString)
		}

		if want.WantParseErrorIs == nil && want.WantParseErrorString == "" && have.ParseError != nil {
			return fmt.Errorf("want clean parse, have error (%v)", have.ParseError)
		}

		if want.WantParseErrorIs != nil && have.ParseError != nil && !errors.Is(have.ParseError, want.WantParseErrorIs) {
			return fmt.Errorf("want wrapped error (%#+v), have error (%#+v)", want.WantParseErrorIs, have.ParseError)
		}

		if want.WantParseErrorString != "" && have.ParseError != nil && !strings.Contains(have.ParseError.Error(), want.WantParseErrorString) {
			return fmt.Errorf("want error string (%q), have error (%v)", want.WantParseErrorString, have.ParseError)
		}

		return nil
	} else {
		if have.ParseError != nil {
			return fmt.Errorf("want no parse error, have error: %v", have.ParseError)
		}
	}

	if want.S != have.S {
		return fmt.Errorf("var S: want %q, have %q", want.S, have.S)
	}
	if want.B != have.B {
		return fmt.Errorf("var B: want %v, have %v", want.B, have.B)
	}
	if want.F != have.F {
		return fmt.Errorf("var F: want %v, have %v", want.F, have.F)
	}
	if !reflect.DeepEqual(want.I, have.I) {
		return fmt.Errorf("var I: want %v, have %v", want.I, have.I)
	}
	if !reflect.DeepEqual(want.Args, have.Args) {
		return fmt.Errorf("var Args: want %v, have %v", want.Args, have.Args)
	}

	return nil
}

func TestBasics(t *testing.T) {
	t.Parallel()

	for _, testcase := range []struct {
		name string
		env  map[string]string
		file string
		args []string
		opts []ff.Option
		want Vars
	}{
		{
			name: "empty",
			args: []string{},
			want: Vars{},
		},
		{
			name: "long flags",
			args: []string{"--string=foo", "--bool=true", "--int32=123", "--int32=456"},
			want: Vars{S: "foo", B: true, I: []int32{123, 456}},
		},
		{
			name: "short and long flags",
			args: []string{"--string=foo", "-s", "bar", "-b", "--int32=1", "-i", "2"},
			want: Vars{S: "bar", B: true, I: []int32{1, 2}},
		},
		{
			name: "flags interspersed with args",
			args: []string{"hello world", "--string=foo", "-b", "another", "argument", "-i", "1"},
			want: Vars{S: "foo", B: true, I: []int32{1}, Args: []string{"hello world", "another", "argument"}},
		},
		{
			name: "args after delimiter",
			args: []string{"--string=foo", "--", "--string=bar"},
			want: Vars{S: "foo", Args: []string{"--string=bar"}},
		},
		{
			name: "config file with mixed prefixes",
			file: "testdata/1.conf",
			opts: []ff.Option{ff.WithConfigFileParser(ff.PlainParser)},
			want: Vars{S: "B", F: 1.23, I: []int32{1, 2, 3, 4}},
		},
		{
			name: "env vars",
			env:  map[string]string{"PF_STRING": "hello", "PF_F": "0.123"},
			opts: []ff.Option{ff.WithEnvVarPrefix("PF")},
			want: Vars{S: "hello", F: 0.123},
		},
		{
			name: "JSON config file",
			file: "testdata/2.json",
			opts: []ff.Option{ff.WithConfigFileParser(ff.JSONParser)},
			want: Vars{S: "hello", I: []int32{1, 2, 3}},
		},
		{
			name: "TOML config file",
			file: "testdata/3.toml",
			opts: []ff.Option{ff.WithConfigFileParser(fftoml.Parser)},
			want: Vars{B: true, I: []int32{50, 100, 150, 0}},
		},
		{
			name: "YAML config file",
			file: "testdata/4.yaml",
			opts: []ff.Option{ff.WithConfigFileParser(ffyaml.Parser)},
			want: Vars{S: "c", F: 123.4, I: []int32{10, 11, 12}},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			if testcase.file != "" {
				testcase.opts = append(testcase.opts, ff.WithConfigFile(testcase.file))
			}

			if len(testcase.env) > 0 {
				for k, v := range testcase.env {
					defer os.Setenv(k, os.Getenv(k))
					os.Setenv(k, v)
				}
			}

			fs, vars := Pair()
			vars.ParseError = ff.Parse(ffpflag.NewFlagSet(fs), testcase.args, testcase.opts...)
			vars.Args = fs.Args()
			if err := Compare(&testcase.want, vars); err != nil {
				t.Fatal(err)
			}
		})
	}
}
