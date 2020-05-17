package ffpflag

import (
	"flag"

	"github.com/spf13/pflag"
)

// FlagSet is an adapter that makes a pflag.FlagSet usable as a ff.FlagSet.
// Flags declared using the P-suffixed variations of the pflag declaration
// functions, with both short and long names, must be referred to using their
// long names only in config files and environment variables.
type FlagSet struct {
	*pflag.FlagSet
}

// NewFlagSet adapts the pflag.FlagSet to a ff.FlagSet.
func NewFlagSet(fs *pflag.FlagSet) *FlagSet {
	return &FlagSet{fs}
}

// Parse implements ff.FlagSet. It calls pflag.FlagSet.Parse directly.
func (fs *FlagSet) Parse(arguments []string) error {
	return fs.FlagSet.Parse(arguments)
}

// Visit implements ff.FlagSet. The flag.Flag provided to the passed function is
// a temporary concrete type constructed from the pflag.Flag.
func (fs *FlagSet) Visit(fn func(*flag.Flag)) {
	fs.FlagSet.Visit(func(pf *pflag.Flag) {
		fn(pflag2std(pf))
	})
}

// VisitAll implements ff.FlagSet. The flag.Flag provided to the passed function
// is a temporary concrete type constructed from the pflag.Flag.
func (fs *FlagSet) VisitAll(fn func(*flag.Flag)) {
	fs.FlagSet.VisitAll(func(pf *pflag.Flag) {
		fn(pflag2std(pf))
	})
}

// Set implements ff.FlagSet. It calls pflag.FlagSet.Set directly.
func (fs *FlagSet) Set(name, value string) error {
	return fs.FlagSet.Set(name, value)
}

// Lookup implements ff.FlagSet. The returned flag.Flag is a temporary concrete
// type constructed from the pflag.Flag.
func (fs *FlagSet) Lookup(name string) *flag.Flag {
	return pflag2std(fs.FlagSet.Lookup(name))
}

func pflag2std(pFlag *pflag.Flag) *flag.Flag {
	if pFlag == nil {
		return nil
	}

	return &flag.Flag{
		Name:     pFlag.Name,
		Usage:    pFlag.Usage,
		Value:    pFlag.Value,
		DefValue: pFlag.DefValue,
	}
}
