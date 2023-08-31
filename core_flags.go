package ff

import (
	"flag"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/peterbourgon/ff/v4/ffval"
)

// CoreFlags is the default implementation of a [Flags]. It's broadly similar to
// a [flag.FlagSet], but with additional capabilities inspired by getopt(3).
//
// CoreFlags is not safe for concurrent use by multiple goroutines.
type CoreFlags struct {
	setName       string
	flagSlice     []*coreFlag
	isParsed      bool
	postParseArgs []string
	isStdAdapter  bool // stdlib package flag behavior: treat -foo the same as --foo
	parent        *CoreFlags
}

var _ Flags = (*CoreFlags)(nil)
var _ Resetter = (*CoreFlags)(nil)

// NewFlags returns a new core flag set with the given name.
func NewFlags(name string) *CoreFlags {
	return &CoreFlags{
		setName:       name,
		flagSlice:     []*coreFlag{},
		isParsed:      false,
		postParseArgs: []string{},
		isStdAdapter:  false,
		parent:        nil,
	}
}

// NewStdFlags returns a core flag set which acts as an adapter for the provided
// [flag.FlagSet], allowing it to implement the [Flags] interface.
//
// The returned core flag set has slightly different behavior than normal. It's
// a fixed "snapshot" of the provided stdfs, which means it doesn't allow new
// flags to be defined, and won't reflect changes made to the stdfs in the
// future. Also, to approximate standard parsing behavior, it treats every flag
// name as a long name, and treats "-" and "--" equivalently when parsing
// arguments.
func NewStdFlags(stdfs *flag.FlagSet) *CoreFlags {
	corefs := NewFlags(stdfs.Name())
	stdfs.VisitAll(func(f *flag.Flag) {
		if _, err := corefs.AddFlag(CoreFlagConfig{
			LongName: f.Name,
			Usage:    f.Usage,
			Value:    f.Value,
		}); err != nil {
			panic(fmt.Errorf("add %s: %w", f.Name, err))
		}
	})
	corefs.isStdAdapter = true
	return corefs
}

// SetParent assigns a parent flag set to this one. In this case, all of the
// flags in all parent flag sets are available, recursively, to the child. For
// example, Parse will match against any parent flag, WalkFlags will traverse
// all parent flags, etc.
//
// This method returns its receiver to allow for builder-style initialization.
func (fs *CoreFlags) SetParent(parent *CoreFlags) *CoreFlags {
	fs.parent = parent
	return fs
}

// GetName returns the name of the flag set provided during construction.
func (fs *CoreFlags) GetName() string {
	return fs.setName
}

// Parse the provided args against the flag set, assigning flag values as
// appropriate. Args are matched to flags defined in this flag set, and, if a
// parent is set, all parent flag sets, recursively. If a specified flag can't
// be found, parse fails with [ErrUnknownFlag]. After a successful parse,
// subsequent calls to parse fail with [ErrAlreadyParsed], until and unless the
// flag set is reset.
func (fs *CoreFlags) Parse(args []string) error {
	if fs.isParsed {
		return ErrAlreadyParsed
	}

	err := fs.parseArgs(args)
	switch {
	case err == nil:
		fs.isParsed = true
	case err != nil:
		fs.postParseArgs = []string{}
	}
	return err
}

func (fs *CoreFlags) parseArgs(args []string) error {
	// Credit where credit is due: this implementation is adapted from
	// https://pkg.go.dev/github.com/pborman/getopt/v2.

	fs.postParseArgs = args

	for len(args) > 0 {
		arg := args[0]
		args = args[1:]

		var (
			isEmpty   = arg == ""
			noDash    = !isEmpty && arg[0] != '-'
			parseDone = isEmpty || noDash
		)
		if parseDone {
			return nil // fs.postParseArgs should include arg
		}

		if arg == "--" {
			fs.postParseArgs = args // fs.postParseArgs should not include "--"
			return nil
		}

		var (
			isLongFlag  = len(arg) > 2 && arg[0:2] == "--"
			isShortFlag = len(arg) > 1 && arg[0] == '-' && !isLongFlag
		)

		// The stdlib package flag parses -abc and --abc the same. If we want to
		// reproduce that behavior, convert -short flags to --long flags. This
		// changes the semantics of concatenated short flags like -abc.
		if isShortFlag && fs.isStdAdapter {
			isShortFlag = false
			isLongFlag = true
			arg = "-" + arg
		}

		var parseErr error
		switch {
		case isShortFlag:
			args, parseErr = fs.parseShortFlag(arg, args)
		case isLongFlag:
			args, parseErr = fs.parseLongFlag(arg, args)
		}
		if parseErr != nil {
			return parseErr
		}

		fs.postParseArgs = args // we parsed arg, so update fs.postParseArgs with the remainder
	}

	return nil
}

// findFlag finds the first matching flag in the flags hierarchy.
func (fs *CoreFlags) findFlag(short rune, long string) *coreFlag {
	var (
		haveShort = isValidShortName(short)
		haveLong  = isValidLongName(long)
	)
	for cursor := fs; cursor != nil; cursor = cursor.parent {
		for _, candidate := range cursor.flagSlice {
			if haveShort && isValidShortName(candidate.shortName) && candidate.shortName == short {
				return candidate
			}
			if haveLong && isValidLongName(candidate.longName) && candidate.longName == long {
				return candidate
			}
		}
	}
	return nil
}

func (fs *CoreFlags) findShortFlag(short rune) *coreFlag {
	return fs.findFlag(short, "")
}

func (fs *CoreFlags) findLongFlag(long string) *coreFlag {
	return fs.findFlag(0, long)
}

func (fs *CoreFlags) parseShortFlag(arg string, args []string) ([]string, error) {
	arg = strings.TrimPrefix(arg, "-")

	for i, r := range arg {
		f := fs.findShortFlag(r)
		if f == nil {
			switch {
			case arg == "-": // `-` == `--`
				return args, nil
			case r == 'h':
				return args, ErrHelp
			default:
				return args, fmt.Errorf("%w %q", ErrUnknownFlag, string(r))
			}
		}

		var value string
		switch {
		case f.isBoolFlag:
			value = "true" // -b -> b=true
		default:
			value = arg[i+1:] // -sabc -> s=abc
			if value == "" {
				if len(args) == 0 {
					return args, newFlagError(f, fmt.Errorf("set: missing argument"))
				}
				value = args[0] // -s abc -> s=abc
				args = args[1:]
			}
		}

		if err := f.flagValue.Set(value); err != nil {
			return args, newFlagError(f, fmt.Errorf("set %q: %w", value, err))
		}
		f.isSet = true

		if !f.isBoolFlag {
			return args, nil
		}
	}

	return args, nil
}

func (fs *CoreFlags) parseLongFlag(arg string, args []string) ([]string, error) {
	var (
		name  string
		value string
	)

	if equals := strings.IndexRune(arg, '='); equals > 0 {
		arg, value = arg[:equals], arg[equals+1:]
	}

	name = strings.TrimPrefix(arg, "--")

	f := fs.findLongFlag(name)
	if f == nil {
		switch {
		case strings.EqualFold(name, "help"):
			return nil, ErrHelp
		case fs.isStdAdapter && strings.EqualFold(name, "h"):
			return nil, ErrHelp
		default:
			return nil, fmt.Errorf("%w %q", ErrUnknownFlag, name)
		}
	}

	if value == "" {
		switch {
		case f.isBoolFlag:
			value = "true" // `-b` or `--foo` default to true
			if len(args) > 0 {
				if _, err := strconv.ParseBool(args[0]); err == nil {
					value = args[0] // `-b true` or `--foo false` should also work
					args = args[1:]
				}
			}
		case !f.isBoolFlag && len(args) > 0:
			value, args = args[0], args[1:]
		case !f.isBoolFlag && len(args) <= 0:
			return nil, fmt.Errorf("missing value")
		default:
			panic("unreachable")
		}
	}

	if err := f.flagValue.Set(value); err != nil {
		return nil, newFlagError(f, fmt.Errorf("set %q: %w", value, err))
	}
	f.isSet = true

	return args, nil
}

// IsParsed returns true if the flag set has been successfully parsed.
func (fs *CoreFlags) IsParsed() bool {
	return fs.isParsed
}

// WalkFlags calls fn for every flag known to the flag set. This includes all
// parent flags, if a parent has been set.
func (fs *CoreFlags) WalkFlags(fn func(Flag) error) error {
	for cursor := fs; cursor != nil; cursor = cursor.parent {
		for _, f := range cursor.flagSlice {
			if err := fn(f); err != nil {
				return err
			}
		}
	}
	return nil
}

// GetFlag returns the first flag known to the flag set that matches the given
// name. This includes all parent flags, if a parent has been set. The name is
// compared against each flag's long name, and, if the name is a single rune,
// it's also compared against each flag's short name.
func (fs *CoreFlags) GetFlag(name string) (Flag, bool) {
	if name == "" {
		return nil, false
	}

	var (
		short = rune(0)
		long  = name
	)
	if len(name) == 1 {
		short, _ = utf8.DecodeRuneInString(name)
	}

	f := fs.findFlag(short, long)
	if f == nil {
		return nil, false
	}

	return f, true
}

// GetArgs returns the args left over after a successful parse.
func (fs *CoreFlags) GetArgs() []string {
	return fs.postParseArgs
}

// Reset the flag set, and all of the flags defined in the flag set, to their
// initial state. After a successful reset, the flag set may be parsed as if it
// were newly constructed.
func (fs *CoreFlags) Reset() error {
	for _, f := range fs.flagSlice {
		if err := f.Reset(); err != nil {
			return newFlagError(f, err)
		}
	}

	fs.postParseArgs = fs.postParseArgs[:0]
	fs.isParsed = false

	return nil
}

// CoreFlagConfig collects the required config for a flag in a core flag set.
type CoreFlagConfig struct {
	// ShortName is the short form name of the flag, which can be provided as a
	// commandline argument with a single dash - prefix. A rune value of 0 or
	// utf8.RuneError is considered an invalid short name and is ignored.
	//
	// At least one of ShortName and/or LongName is required.
	ShortName rune

	// LongName is the long form name of the flag, which can be provided as a
	// commandline argument with a double-dash -- prefix. An empty string is
	// considered an invalid long name and is ignored.
	//
	// At least one of ShortName and/or LongName is required.
	LongName string

	// Placeholder is typically used to represent an example value in the help
	// text for the flag. For example, a placeholder of `BAR` might result in
	// help text like
	//
	//      -f, --foo BAR   set the foo parameter
	//
	// The placeholder is determined by the following logic.
	//
	//  - If EmptyPlaceholder is true, use the empty string
	//  - If Placeholder is non-empty, use that string
	//  - If Usage contains a `backtick-quoted` substring, use that substring
	//  - If Value is a boolean with default value false, use the empty string
	//  - Otherwise, use a simple transformation of the concrete Value type name
	//
	// Optional.
	Placeholder string

	// NoPlaceholder forces the placeholder of the flag to the empty string.
	// This can be useful if you want to elide the placeholder from help text.
	NoPlaceholder bool

	// Usage is a short help message for the flag, typically printed after the
	// flag name(s) on a single line in the help output. For example, a foo flag
	// might have the usage string "set the foo parameter", which might be
	// rendered as follows.
	//
	//      -f, --foo BAR   set the foo parameter
	//
	// If the usage string contains a `backtick` quoted substring, that
	// substring will be treated as a placeholder, if a placeholder was not
	// otherwise explicitly provided.
	//
	// Recommended.
	Usage string

	// Value is used to parse and store the actual flag value. The MakeFlagValue
	// helper can be used to construct values for common primitive types.
	//
	// As a special case, if the value has an IsBoolFlag() bool method returning
	// true, then it will be treated as a boolean flag. Boolean flags are parsed
	// slightly differently than normal flags: they can be provided without an
	// explicit value, in which case the value is assumed to be true.
	//
	// Required.
	Value flag.Value

	// NoDefault forces the default value of the flag to the empty string. This
	// can be useful if you want to elide the default value from help text.
	NoDefault bool
}

func (cfg CoreFlagConfig) isBoolFlag() bool {
	bf, ok := cfg.Value.(IsBoolFlagger)
	if !ok {
		return false
	}
	return bf.IsBoolFlag()
}

func (cfg CoreFlagConfig) getPlaceholder() string {
	// If a placeholder is explicitly refused, use an empty string.
	if cfg.NoPlaceholder {
		return ""
	}

	// If a placeholder is explicitly provided, use that.
	if cfg.Placeholder != "" {
		return cfg.Placeholder
	}

	// If the usage text contains a `backticked` substring, use that.
	for i := 0; i < len(cfg.Usage); i++ {
		if cfg.Usage[i] == '`' {
			for j := i + 1; j < len(cfg.Usage); j++ {
				if cfg.Usage[j] == '`' {
					return cfg.Usage[i+1 : j]
				}
			}
			break
		}
	}

	// Bool flags with default value false should have empty placeholders.
	if bf, ok := cfg.Value.(IsBoolFlagger); ok && bf.IsBoolFlag() {
		if b, err := strconv.ParseBool(cfg.Value.String()); err == nil && !b {
			return ""
		}
	}

	// Otherwise, use a transformation of the flag value type name.
	var typeName string
	typeName = fmt.Sprintf("%T", cfg.Value)
	typeName = strings.ToUpper(typeName)
	typeName = typeNameDefaultRegexp.ReplaceAllString(typeName, "$1")
	typeName = strings.TrimSuffix(typeName, "VALUE")
	if lastDot := strings.LastIndex(typeName, "."); lastDot > 0 {
		typeName = typeName[lastDot+1:]
	}
	return typeName
}

var typeNameDefaultRegexp = regexp.MustCompile(`[A-Z0-9\_\.\*]+\[(.+)\]`)

func (cfg CoreFlagConfig) getDefaultValue() string {
	// If the config explicitly declares an empty default, use the empty string.
	if cfg.NoDefault {
		return ""
	}

	// Otherwise, use Value.String.
	return cfg.Value.String()
}

// AddFlag adds a flag to the flag set, as specified by the provided config. An
// error is returned if the config is invalid, or if a flag is already defined
// in the flag set with the same short or long name.
//
// This is a fairly low level method. Consumers may prefer type-specific helpers
// like [CoreFlags.Bool], [CoreFlags.StringVar], etc.
func (fs *CoreFlags) AddFlag(cfg CoreFlagConfig) (Flag, error) {
	if fs.isStdAdapter {
		return nil, fmt.Errorf("cannot add flags to standard flag set adapter")
	}

	if cfg.Value == nil {
		return nil, fmt.Errorf("value is required")
	}

	var (
		hasShort     = isValidShortName(cfg.ShortName)
		hasLong      = isValidLongName(cfg.LongName)
		isBoolFlag   = cfg.isBoolFlag()
		placeholder  = cfg.getPlaceholder()
		defaultValue = cfg.getDefaultValue()
	)
	if !hasShort && !hasLong {
		return nil, fmt.Errorf("short name and/or long name is required")
	}
	if isBoolFlag && !hasLong {
		if b, err := strconv.ParseBool(defaultValue); err == nil && b {
			return nil, fmt.Errorf("%s: default true boolean flag requires a long name", string(cfg.ShortName))
		}
	}

	f := &coreFlag{
		flagSet:     fs,
		shortName:   cfg.ShortName,
		longName:    cfg.LongName,
		placeholder: placeholder,
		defaultval:  defaultValue,
		usageval:    cfg.Usage,
		flagValue:   cfg.Value,
		isBoolFlag:  isBoolFlag,
	}

	for _, existing := range fs.flagSlice {
		if isDuplicate(f, existing) {
			return nil, newFlagError(f, ErrDuplicateFlag)
		}
	}

	fs.flagSlice = append(fs.flagSlice, f)

	return f, nil
}

// Value defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlags) Value(short rune, long string, value flag.Value, usage string) Flag {
	f, err := fs.AddFlag(CoreFlagConfig{
		ShortName: short,
		LongName:  long,
		Usage:     usage,
		Value:     value,
	})
	if err != nil {
		panic(err)
	}
	return f
}

// ValueShort defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlags) ValueShort(short rune, value flag.Value, usage string) Flag {
	return fs.Value(short, "", value, usage)
}

// ValueLong defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlags) ValueLong(long string, value flag.Value, usage string) Flag {
	return fs.Value(0, long, value, usage)
}

// BoolVar defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlags) BoolVar(pointer *bool, short rune, long string, def bool, usage string) Flag {
	return fs.Value(short, long, ffval.NewValueDefault(pointer, def), usage)
}

// Bool defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlags) Bool(short rune, long string, def bool, usage string) *bool {
	var value bool
	fs.BoolVar(&value, short, long, def, usage)
	return &value
}

// BoolShort defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlags) BoolShort(short rune, def bool, usage string) *bool {
	return fs.Bool(short, "", def, usage)
}

// BoolLong defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlags) BoolLong(long string, def bool, usage string) *bool {
	return fs.Bool(0, long, def, usage)
}

// StringVar defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlags) StringVar(pointer *string, short rune, long string, def string, usage string) Flag {
	return fs.Value(short, long, ffval.NewValueDefault(pointer, def), usage)
}

// String defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlags) String(short rune, long string, def string, usage string) *string {
	var value string
	fs.StringVar(&value, short, long, def, usage)
	return &value
}

// StringShort defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlags) StringShort(short rune, def string, usage string) *string {
	return fs.String(short, "", def, usage)
}

// StringLong defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlags) StringLong(long string, def string, usage string) *string {
	return fs.String(0, long, def, usage)
}

// StringListVar defines a new flag in the flag set, and panics on any error.
//
// The flag represents a list of strings, where each call to Set adds a new
// value to the list. Duplicate values are permitted.
func (fs *CoreFlags) StringListVar(pointer *[]string, short rune, long string, usage string) Flag {
	return fs.Value(short, long, ffval.NewList(pointer), usage)
}

// StringList defines a new flag in the flag set, and panics on any error.
// See [StringListVar] for more details.
func (fs *CoreFlags) StringList(short rune, long string, usage string) *[]string {
	var value []string
	fs.StringListVar(&value, short, long, usage)
	return &value
}

// StringListShort defines a new flag in the flag set, and panics on any error.
// See [StringListVar] for more details.
func (fs *CoreFlags) StringListShort(short rune, usage string) *[]string {
	return fs.StringList(short, "", usage)
}

// StringListLong defines a new flag in the flag set, and panics on any error.
// See [StringListVar] for more details.
func (fs *CoreFlags) StringListLong(long string, usage string) *[]string {
	return fs.StringList(0, long, usage)
}

// StringSetVar defines a new flag in the flag set, and panics on any error.
//
// The flag represents a unique list of strings, where each call to Set adds a
// new value to the list. Duplicate values are silently dropped.
func (fs *CoreFlags) StringSetVar(pointer *[]string, short rune, long string, usage string) Flag {
	return fs.Value(short, long, ffval.NewUniqueList(pointer), usage)
}

// StringSet defines a new flag in the flag set, and panics on any error.
// See [StringSetVar] for more details.
func (fs *CoreFlags) StringSet(short rune, long string, usage string) *[]string {
	var value []string
	fs.StringSetVar(&value, short, long, usage)
	return &value
}

// StringSetShort defines a new flag in the flag set, and panics on any error.
// See [StringSetVar] for more details.
func (fs *CoreFlags) StringSetShort(short rune, usage string) *[]string {
	return fs.StringSet(short, "", usage)
}

// StringSetLong defines a new flag in the flag set, and panics on any error.
// See [StringSetVar] for more details.
func (fs *CoreFlags) StringSetLong(long string, usage string) *[]string {
	return fs.StringSet(0, long, usage)
}

// Float64Var defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlags) Float64Var(pointer *float64, short rune, long string, def float64, usage string) Flag {
	return fs.Value(short, long, ffval.NewValueDefault(pointer, def), usage)
}

// Float64 defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlags) Float64(short rune, long string, def float64, usage string) *float64 {
	var value float64
	fs.Float64Var(&value, short, long, def, usage)
	return &value
}

// Float64Short defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlags) Float64Short(short rune, def float64, usage string) *float64 {
	return fs.Float64(short, "", def, usage)
}

// Float64Long defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlags) Float64Long(long string, def float64, usage string) *float64 {
	return fs.Float64(0, long, def, usage)
}

// IntVar defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlags) IntVar(pointer *int, short rune, long string, def int, usage string) Flag {
	return fs.Value(short, long, ffval.NewValueDefault(pointer, def), usage)
}

// Int defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlags) Int(short rune, long string, def int, usage string) *int {
	var value int
	fs.IntVar(&value, short, long, def, usage)
	return &value
}

// IntShort defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlags) IntShort(short rune, def int, usage string) *int {
	return fs.Int(short, "", def, usage)
}

// IntLong defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlags) IntLong(long string, def int, usage string) *int {
	return fs.Int(0, long, def, usage)
}

// UintVar defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlags) UintVar(pointer *uint, short rune, long string, def uint, usage string) Flag {
	return fs.Value(short, long, ffval.NewValueDefault(pointer, def), usage)
}

// Uint defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlags) Uint(short rune, long string, def uint, usage string) *uint {
	var value uint
	fs.UintVar(&value, short, long, def, usage)
	return &value
}

// UintShort defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlags) UintShort(short rune, def uint, usage string) *uint {
	return fs.Uint(short, "", def, usage)
}

// UintLong defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlags) UintLong(long string, def uint, usage string) *uint {
	return fs.Uint(0, long, def, usage)
}

// Uint64Var defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlags) Uint64Var(pointer *uint64, short rune, long string, def uint64, usage string) Flag {
	return fs.Value(short, long, ffval.NewValueDefault(pointer, def), usage)
}

// Uint64 defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlags) Uint64(short rune, long string, def uint64, usage string) *uint64 {
	var value uint64
	fs.Uint64Var(&value, short, long, def, usage)
	return &value
}

// Uint64Short defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlags) Uint64Short(short rune, def uint64, usage string) *uint64 {
	return fs.Uint64(short, "", def, usage)
}

// Uint64Long defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlags) Uint64Long(long string, def uint64, usage string) *uint64 {
	return fs.Uint64(0, long, def, usage)
}

// DurationVar defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlags) DurationVar(pointer *time.Duration, short rune, long string, def time.Duration, usage string) Flag {
	return fs.Value(short, long, ffval.NewValueDefault(pointer, def), usage)
}

// Duration defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlags) Duration(short rune, long string, def time.Duration, usage string) *time.Duration {
	var value time.Duration
	fs.DurationVar(&value, short, long, def, usage)
	return &value
}

// DurationShort defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlags) DurationShort(short rune, def time.Duration, usage string) *time.Duration {
	return fs.Duration(short, "", def, usage)
}

// DurationLong defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlags) DurationLong(long string, def time.Duration, usage string) *time.Duration {
	return fs.Duration(0, long, def, usage)
}

// Func defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlags) Func(short rune, long string, fn func(string) error, usage string) {
	stdfs := flag.NewFlagSet("flagset-name", flag.ContinueOnError)
	stdfs.Func("flag-name", "flag-usage", fn)
	value := stdfs.Lookup("flag-name").Value
	fs.Value(short, long, value, usage)
}

// FuncShort defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlags) FuncShort(short rune, fn func(string) error, usage string) {
	fs.Func(short, "", fn, usage)
}

// FuncLong defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlags) FuncLong(long string, fn func(string) error, usage string) {
	fs.Func(0, long, fn, usage)
}

//
//
//

type coreFlag struct {
	flagSet     *CoreFlags
	shortName   rune
	longName    string
	placeholder string
	defaultval  string
	usageval    string
	flagValue   flag.Value
	isBoolFlag  bool
	isSet       bool
}

var _ Flag = (*coreFlag)(nil)

func (f *coreFlag) GetFlags() Flags {
	return f.flagSet
}

func (f *coreFlag) GetShortName() (rune, bool) {
	return f.shortName, isValidShortName(f.shortName)
}

func (f *coreFlag) GetLongName() (string, bool) {
	return f.longName, isValidLongName(f.longName)
}

func (f *coreFlag) GetPlaceholder() string {
	return f.placeholder
}

func (f *coreFlag) GetDefault() string {
	return f.defaultval
}

func (f *coreFlag) GetUsage() string {
	return f.usageval
}

func (f *coreFlag) SetValue(s string) error {
	if err := f.flagValue.Set(s); err != nil {
		return err
	}
	f.isSet = true
	return nil
}

func (f *coreFlag) GetValue() string {
	return f.flagValue.String()
}

func (f *coreFlag) IsSet() bool {
	return f.isSet
}

func (f *coreFlag) Reset() error {
	if r, ok := f.flagValue.(Resetter); ok {
		if err := r.Reset(); err != nil {
			return err
		}
	}

	if err := f.flagValue.Set(f.defaultval); err != nil {
		return err
	}
	f.isSet = false

	return nil
}

func (f *coreFlag) IsStdFlag() bool {
	return f.flagSet.isStdAdapter
}

func isDuplicate(incoming, existing *coreFlag) bool {
	var (
		sameShortName = isValidShortName(incoming.shortName) && isValidShortName(existing.shortName) && incoming.shortName == existing.shortName
		sameLongName  = isValidLongName(incoming.longName) && isValidLongName(existing.longName) && incoming.longName == existing.longName
		shortIsLong   = isValidShortName(incoming.shortName) && isValidLongName(existing.longName) && len(existing.longName) == 1 && string(incoming.shortName) == existing.longName
		longIsShort   = isValidLongName(incoming.longName) && isValidShortName(existing.shortName) && len(incoming.longName) == 1 && incoming.longName == string(existing.shortName)
		isDuplicate   = sameShortName || sameLongName || shortIsLong || longIsShort
	)
	return isDuplicate
}
