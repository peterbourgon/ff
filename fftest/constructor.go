package fftest

import (
	"flag"
	"fmt"
	"strings"

	"github.com/peterbourgon/ff/v4"
)

// Constructor produces a flag set, and a set of vars managed by that flag set.
type Constructor struct {
	// Name of the constructor, used in test names.
	Name string

	// Make should return a flag set and vars structure, where each value in the
	// vars structure is updated by a corresponding flag in the flag set. The
	// default value for each flag should be taken from the def parameter.
	Make func(def Vars) (ff.Flags, *Vars)
}

// CoreConstructor produces a core flag set, with both short and long flag names
// for each value.
var CoreConstructor = Constructor{
	Name: "core",
	Make: func(def Vars) (ff.Flags, *Vars) {
		var v Vars
		fs := ff.NewFlags("fftest")
		fs.StringVar(&v.S, 's', "str", def.S, "string")
		fs.IntVar(&v.I, 'i', "int", def.I, "int")
		fs.Float64Var(&v.F, 'f', "flt", def.F, "float64")
		fs.BoolVar(&v.A, 'a', "aflag", def.A, "bool a")
		fs.BoolVar(&v.B, 'b', "bflag", def.B, "bool b")
		fs.BoolVar(&v.C, 'c', "cflag", def.C, "bool c")
		fs.DurationVar(&v.D, 'd', "dur", def.D, "time.Duration")
		fs.AddFlag(ff.CoreFlagConfig{ShortName: 'x', LongName: "xxx", Placeholder: "STR", Usage: "collection of strings (repeatable)", Value: &v.X})
		return fs, &v
	},
}

// StdConstructor produces a stdlib flag set adapter.
var StdConstructor = Constructor{
	Name: "std",
	Make: func(def Vars) (ff.Flags, *Vars) {
		var v Vars
		fs := flag.NewFlagSet("fftest", flag.ContinueOnError)
		fs.StringVar(&v.S, "s", def.S, "string")
		fs.IntVar(&v.I, "i", def.I, "int")
		fs.Float64Var(&v.F, "f", def.F, "float64")
		fs.BoolVar(&v.A, "a", def.A, "bool a")
		fs.BoolVar(&v.B, "b", def.B, "bool b")
		fs.BoolVar(&v.C, "c", def.C, "bool c")
		fs.DurationVar(&v.D, "d", def.D, "time.Duration")
		fs.Var(&v.X, "x", "collection of strings (repeatable)")
		return ff.NewStdFlags(fs), &v
	},
}

// DefaultConstructors are used for test cases that don't specify constructors.
var DefaultConstructors = []Constructor{
	CoreConstructor,
	StdConstructor,
}

// NewNestedConstructor returns a constructor where flags have specific
// hierarchical names delimited by the provided delim. This is useful for
// testing config file formats that allow nested configuration.
func NewNestedConstructor(delim string) Constructor {
	return Constructor{
		Name: fmt.Sprintf("nested delimiter '%s'", delim),
		Make: func(def Vars) (ff.Flags, *Vars) {
			var (
				skey = strings.Join([]string{"foo", "bar", "s"}, delim)
				ikey = strings.Join([]string{"nested", "i"}, delim)
				fkey = strings.Join([]string{"nested", "f"}, delim)
				akey = strings.Join([]string{"nested", "a"}, delim)
				bkey = strings.Join([]string{"nested", "b"}, delim)
				ckey = strings.Join([]string{"nested", "c"}, delim)
				xkey = strings.Join([]string{"x", "value"}, delim)
			)
			var v Vars
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			fs.StringVar(&v.S, skey, def.S, "string var")
			fs.IntVar(&v.I, ikey, def.I, "int var")
			fs.Float64Var(&v.F, fkey, def.F, "float64 var")
			fs.BoolVar(&v.A, akey, def.A, "bool var a")
			fs.BoolVar(&v.B, bkey, def.B, "bool var b")
			fs.BoolVar(&v.C, ckey, def.C, "bool var c")
			fs.Var(&v.X, xkey, "x var")
			return ff.NewStdFlags(fs), &v
		},
	}
}
