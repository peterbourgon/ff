package ff

import (
	"flag"

	"github.com/peterbourgon/ff/v4/ffval"
)

// Flags describes a collection of flags, typically associated with a specific
// command (or sub-command) executed by an end user.
//
// Any valid Flags can be provided to [Parse], or used as the Flags field in a
// [Command]. This allows custom flag set implementations to take advantage of
// the primary features of this module.
//
// Implementations are not expected to be safe for concurrent use by multiple
// goroutines.
type Flags interface {
	// GetName should return the name of the flag set.
	GetName() string

	// Parse should parse the provided args against the flag set, setting flags
	// as appropriate, and saving leftover args to be returned by GetArgs. The
	// provided args shouldn't include the program name: callers should pass
	// os.Args[1:], not os.Args.
	Parse(args []string) error

	// IsParsed should return true if the flag set was successfully parsed.
	IsParsed() bool

	// WalkFlags should call the given fn for each flag known to the flag set.
	// Note that this may include flags that are actually defined in different
	// "parent" flag sets. If fn returns an error, WalkFlags should immediately
	// return that error.
	WalkFlags(fn func(Flag) error) error

	// GetFlag should find and return the first flag known to the flag set with
	// the given name. The name should always be compared against valid flag
	// long names. If name is a single valid rune, it should also be compared
	// against valid flag short names. Note that this may return a flag that is
	// actually defined in a different "parent" flag set.
	GetFlag(name string) (Flag, bool)

	// GetArgs should return the args left over after a successful call to
	// parse. If parse has not yet been called successfully, it should return an
	// empty (or nil) slice.
	GetArgs() []string
}

// Flag describes a single runtime configuration parameter, defined within a set
// of flags, and with a value that's parsed from a string.
//
// Implementations are not expected to be safe for concurrent use by multiple
// goroutines.
type Flag interface {
	// GetFlags should return the set of flags in which this flag is defined.
	// It's primarily used for help output.
	GetFlags() Flags

	// GetShortName should return the short name for this flag, if one is
	// defined. A short name is always a single character (rune) which is
	// typically parsed with a single leading hyphen, e.g. -f.
	GetShortName() (rune, bool)

	// GetLongName should return the long name for this flag, if one is defined.
	// A long name is always a non-empty string which is typically parsed with
	// two leading hyphens, e.g. --foo.
	GetLongName() (string, bool)

	// GetPlaceholder should return a string that can be used as a placeholder
	// for the flag value in help output. For example, a placeholder for a
	// string flag might be STRING. An empty placeholder is valid.
	GetPlaceholder() string

	// GetUsage should return a short description of the flag, which can be
	// included in the help output on the same line as the flag name(s). For
	// example, the usage string for a timeout flag used in an HTTP client might
	// be "timeout for outgoing HTTP requests". An empty usage string is valid,
	// but not recommended.
	GetUsage() string

	// GetDefault should return the default value of the flag as a string.
	GetDefault() string

	// SetValue should parse the provided string into the appropriate type for
	// the flag, and set the flag to that parsed value.
	SetValue(string) error

	// GetValue should return the current value of the flag as a string. If no
	// value has been set, it should return the default value as a string.
	GetValue() string

	// IsSet should return true if SetValue has been called successfully.
	IsSet() bool
}

// Resetter may optionally be implemented by [Flags].
type Resetter interface {
	// Reset should revert the flag set to its initial state, including all
	// flags defined in the flag set. If reset returns successfully, the flag
	// set should be as if it were newly constructed: IsParsed should return
	// false, GetArgs should return an empty slice, etc.
	Reset() error
}

// IsBoolFlagger is used to identify flag values representing booleans.
type IsBoolFlagger interface{ IsBoolFlag() bool }

var (
	_ flag.Value    = (*ffval.Value[any])(nil)
	_ Resetter      = (*ffval.Value[any])(nil)
	_ IsBoolFlagger = (*ffval.Value[any])(nil)
)
