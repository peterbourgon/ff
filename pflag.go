package ff

import (
	"flag"

	"github.com/spf13/pflag"
)

// FromPflag adapts pflag.FlagSet to FlagSet interface defined in this package.
func FromPflag(fs *pflag.FlagSet) FlagSet {
	return pFlagSet{base: fs}
}

type pFlagSet struct {
	base *pflag.FlagSet
}

func (fs pFlagSet) Parse(arguments []string) error {
	return fs.base.Parse(arguments)
}

func (fs pFlagSet) Visit(fn func(*flag.Flag)) {
	fs.base.Visit(func(pFlag *pflag.Flag) {
		fn(pflag2std(pFlag))
	})
}

func (fs pFlagSet) VisitAll(fn func(*flag.Flag)) {
	fs.base.VisitAll(func(pFlag *pflag.Flag) {
		fn(pflag2std(pFlag))
	})
}

func (fs pFlagSet) Set(name, value string) error {
	return fs.base.Set(name, value)
}

func (fs pFlagSet) Lookup(name string) *flag.Flag {
	return pflag2std(fs.base.Lookup(name))
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
