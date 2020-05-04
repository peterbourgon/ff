package ff

import (
	"flag"

	"github.com/spf13/pflag"
)

type FlagSet interface {
	Parse([]string) error
	Visit(fn func(name string))
	VisitAll(fn func(name string))
	Set(name string, value string) error
	Lookup(name string) (value string, ok bool)
}

// NewFromFlag adapt flag.FlagSet to FlagSet defined in this packaga.
func NewFromFlag(fs *flag.FlagSet) FlagSet {
	return stdFS{base: fs}
}

type stdFS struct {
	base *flag.FlagSet
}

func (fs stdFS) Parse(arguments []string) error {
	return fs.base.Parse(arguments)
}

func (fs stdFS) Set(name, value string) error {
	return fs.base.Set(name, value)
}

func (fs stdFS) Visit(fn func(name string)) {
	fs.base.Visit(func(f *flag.Flag) {
		fn(f.Name)
	})
}
func (fs stdFS) VisitAll(fn func(name string)) {
	fs.base.VisitAll(func(f *flag.Flag) {
		fn(f.Name)
	})
}

func (fs stdFS) Lookup(s string) (value string, ok bool) {
	f := fs.base.Lookup(s)
	if f != nil {
		return f.Value.String(), true
	}
	return "", false
}

// NewFromPflag adapt pflag.FlagSet to FlagSet defined in this packaga.
func NewFromPflag(fs *pflag.FlagSet) FlagSet {
	return pflagFS{base: fs}
}

type pflagFS struct {
	base *pflag.FlagSet
}

func (fs pflagFS) Parse(arguments []string) error {
	return fs.base.Parse(arguments)
}

func (fs pflagFS) Set(name, value string) error {
	return fs.base.Set(name, value)
}

func (fs pflagFS) Visit(fn func(name string)) {
	fs.base.Visit(func(f *pflag.Flag) {
		fn(f.Name)
	})
}
func (fs pflagFS) VisitAll(fn func(name string)) {
	fs.base.VisitAll(func(f *pflag.Flag) {
		fn(f.Name)
	})
}

func (fs pflagFS) Lookup(s string) (value string, ok bool) {
	f := fs.base.Lookup(s)
	if f != nil {
		return f.Value.String(), true
	}
	return "", false
}
