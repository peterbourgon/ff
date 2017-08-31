package ff

import (
	"errors"
	"fmt"
	"os"
	"text/tabwriter"
	"time"
)

// ErrHelp is returned by Parse when the user passed -h.
var ErrHelp = errors.New("help requested")

// FlagSet is a collection of variables.
type FlagSet struct {
	// Usage prints help text to stdout.
	Usage func()

	short string
	vars  map[string]Variable
}

// NewFlagSet returns an empty set of variables.
// The short help text will be displayed with -h.
func NewFlagSet(short string) *FlagSet {
	fs := &FlagSet{
		short: short,
		vars:  map[string]Variable{},
	}
	fs.Usage = fs.usage
	return fs
}

// String allocates and returns a string variable
// that's been installed to the flag set.
func (fs *FlagSet) String(name string, def string, usage string, options ...VariableOption) *string {
	var v string
	fs.StringVar(&v, name, def, usage, options...)
	return &v
}

// StringVar installs a remote string variable to the flag set.
func (fs *FlagSet) StringVar(p *string, name string, def string, usage string, options ...VariableOption) {
	fs.Var(newStringValue(def, p), name, def, usage, options...)
}

// Duration allocates and returns a time.Duration variable
// that's been installed to the flag set.
func (fs *FlagSet) Duration(name string, def time.Duration, usage string, options ...VariableOption) *time.Duration {
	var v time.Duration
	fs.DurationVar(&v, name, def, usage, options...)
	return &v
}

// DurationVar installs a remote time.Duration variable to the flag set.
func (fs *FlagSet) DurationVar(p *time.Duration, name string, def time.Duration, usage string, options ...VariableOption) {
	fs.Var(newDurationValue(def, p), name, def.String(), usage, options...)
}

// Var installs a variable of any concrete type to the flag set.
func (fs *FlagSet) Var(value Value, name, def, usage string, options ...VariableOption) {
	v := Variable{
		value: value,
		name:  name,
		def:   def,
		usage: usage,
		keys:  map[SourceType]string{},
	}
	for _, option := range options {
		option(&v)
	}
	fs.vars[name] = v
}

// Parse defined variables from commandline flags, and optionally other sources,
// in decreasing order of priority. If error is non-nil, the set must be
// considered invalid.
func (fs *FlagSet) Parse(args []string, otherSources ...Source) error {
	// To support -conf=whatever.json, we need to do a first pass over the
	// defined variables and fill them from commandline flags.
	if err := fs.apply(args); err != nil {
		return err
	}

	// Now we can walk each parse source in priority order.
	for i := len(otherSources) - 1; i >= 0; i-- {
		var (
			src = otherSources[i]
			typ = src.Type()
			dat = src.Fetch(fs)
		)
		for _, v := range fs.vars {
			key, ok := v.keys[typ]
			if !ok {
				continue
			}
			val, ok := dat[key]
			if !ok {
				continue
			}
			if err := v.value.Set(val); err != nil {
				return err
			}
		}
	}

	// Commandline flags are always highest-priority; re-apply them.
	if err := fs.apply(args); err != nil {
		return err
	}

	return nil
}

func (fs *FlagSet) apply(args []string) error {
	var (
		done bool
		err  error
	)
	for {
		args, done, err = fs.parseOne(args)
		if err != nil {
			return err
		}
		if done {
			break
		}
	}
	return nil
}

func (fs *FlagSet) parseOne(inArgs []string) (args []string, done bool, err error) {
	// Adapted from https://golang.org/src/flag/flag.go
	if len(inArgs) == 0 {
		return inArgs, true, nil
	}

	var str string
	str, args = inArgs[0], inArgs
	if len(str) == 0 || str[0] != '-' || len(str) == 1 {
		return inArgs[1:], false, nil // skip and ignore
	}

	numMinuses := 1
	if str[1] == '-' {
		numMinuses++
		if len(str) == 2 { // "--" terminates the flags
			return args[1:], true, nil // done
		}
	}

	name := str[numMinuses:]
	if len(name) == 0 || name[0] == '-' || name[0] == '=' {
		return args, false, fmt.Errorf("bad flag syntax: %s", str)
	}

	// It's a flag. Does it have an argument?
	args = args[1:]
	hasValue := false
	value := ""
	for i := 1; i < len(name); i++ { // equals cannot be first
		if name[i] == '=' {
			value = name[i+1:]
			hasValue = true
			name = name[0:i]
			break
		}
	}

	v, defined := fs.vars[name]
	if !defined {
		if name == "help" || name == "h" { // special case for nice help message
			fs.Usage()
			return args, false, ErrHelp
		}
		return args, false, fmt.Errorf("flag provided but not defined: -%s", name)
	}

	// TODO(pb): special case for bool flags
	//if fv, ok := flag.Value.(boolFlag); ok && fv.IsBoolFlag() { // special case: doesn't need an arg
	//	if hasValue {
	//		if err := fv.Set(value); err != nil {
	//			return "", nil, false, f.failf("invalid boolean value %q for -%s: %v", value, name, err)
	//		}
	//	} else {
	//		if err := fv.Set("true"); err != nil {
	//			return "", nil, false, f.failf("invalid boolean flag %s: %v", name, err)
	//		}
	//	}
	//} else {
	// It must have a value, which might be the next argument.
	if !hasValue && len(args) > 0 {
		// value is the next arg
		hasValue = true
		value, args = args[0], args[1:]
	}
	if !hasValue {
		return args, false, fmt.Errorf("flag needs an argument: -%s", name)
	}
	//}

	if err := v.value.Set(value); err != nil {
		return args, false, fmt.Errorf("invalid value %q for flag -%s: %v", value, name, err)
	}

	return args, false, nil
}

func (fs *FlagSet) usage() {
	fmt.Fprintf(os.Stderr, "USAGE\n")
	fmt.Fprintf(os.Stderr, "  %s\n", fs.short)
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "FLAGS\n")
	w := tabwriter.NewWriter(os.Stderr, 0, 2, 2, ' ', 0)
	for _, v := range fs.vars {
		fmt.Fprintf(w, "\t-%s %s\t%s\n", v.name, v.def, v.usage)
	}
	w.Flush()
	fmt.Fprintf(os.Stderr, "\n")

}
