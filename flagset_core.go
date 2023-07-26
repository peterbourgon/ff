package ff

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// CoreFlagSet is the default implementation of a [FlagSet]. It's largely
// inspired by getopt(3), but is not equivalent to, or compatible with, that
// tool.
//
// Flags may be defined with short (-f) and/or long (--foo) names. Flags use the
// [flag.Value] interface to parse and maintain their underlying data.
//
// A flag set may optionally be assigned a parent. In this case, all of the
// flags in parent flag sets are recursively accessible to the child. Every
// method which interacts with flags is affected. For example, WalkFlags will
// enumerate the complete set of flags; Parse is able to set any of those flags;
// etc. This is useful in commmand hierarchies, so that flags defined in a base
// command are visible, and parsable, by subcommands.
type CoreFlagSet struct {
	flagSetName   string
	flagSlice     []*coreFlag
	isParsed      bool
	postParseArgs []string
	isStdAdapter  bool // stdlib package flag behavior: treat -foo the same as --foo
	parent        *CoreFlagSet
}

var _ FlagSet = (*CoreFlagSet)(nil)

// NewSet returns a new core flag set with the given name.
func NewSet(name string) *CoreFlagSet {
	return &CoreFlagSet{
		flagSetName:   name,
		flagSlice:     []*coreFlag{},
		isParsed:      false,
		postParseArgs: []string{},
		isStdAdapter:  false,
		parent:        nil,
	}
}

// NewStdSet returns a core flag set which acts as an adapter for the provided
// [flag.FlagSet], allowing it to implement the [FlagSet] interface.
//
// The returned core flag set behaves differently than normal. It's a fixed
// "snapshot" of the provided standard flag set, and so doesn't reflect changes
// made to the standard flag set in the future, and doesn't allow new flags to
// be added. Also, it parses every flag name as a long name, even if it has just
// a single leading hyphen, to approximate the standard parsing behavior.
func NewStdSet(stdfs *flag.FlagSet) *CoreFlagSet {
	corefs := NewSet(stdfs.Name())
	stdfs.VisitAll(func(f *flag.Flag) {
		if err := corefs.AddFlag(CoreFlagConfig{
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
// parent flags are available, recursively, to the child. For example, Parse
// will match against any parent flag, WalkFlags will traverse all parent flags,
// etc.
//
// This method returns its receiver to allow for builder-style initialization.
func (fs *CoreFlagSet) SetParent(parent *CoreFlagSet) *CoreFlagSet {
	fs.parent = parent
	return fs
}

// GetFlagSetName returns the name of the flag set provided during construction.
func (fs *CoreFlagSet) GetFlagSetName() string {
	return fs.flagSetName
}

// Parse the provided args against the flag set, assigning flag values as
// appropriate. Args are matched to flags defined in this flag set, and, if a
// parent is set, all parent flag sets, recursively. If a specified flag can't
// be found, parse fails with [ErrUnknownFlag]. After a successful parse,
// subsequent calls to parse fail with [ErrAlreadyParsed], until and unless the
// flag set is reset.
func (fs *CoreFlagSet) Parse(args []string) error {
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

func (fs *CoreFlagSet) parseArgs(args []string) error {
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
		// reproduce that behavior, convert -short flags to --long flags.
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

func (fs *CoreFlagSet) findFlag(short rune, long string) *coreFlag {
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

func (fs *CoreFlagSet) findShortFlag(short rune) *coreFlag {
	return fs.findFlag(short, "")
}

func (fs *CoreFlagSet) findLongFlag(long string) *coreFlag {
	return fs.findFlag(0, long)
}

func (fs *CoreFlagSet) parseShortFlag(arg string, args []string) ([]string, error) {
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

func (fs *CoreFlagSet) parseLongFlag(arg string, args []string) ([]string, error) {
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
func (fs *CoreFlagSet) IsParsed() bool {
	return fs.isParsed
}

// WalkFlags calls fn for every flag known to the flag set. This includes all
// parent flags, if a parent has been set.
func (fs *CoreFlagSet) WalkFlags(fn func(Flag) error) error {
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
func (fs *CoreFlagSet) GetFlag(name string) (Flag, bool) {
	if name == "" {
		return nil, false
	}

	var (
		short, _ = utf8.DecodeRuneInString(name)
		long     = name
	)
	f := fs.findFlag(short, long)
	if f == nil {
		return nil, false
	}

	return f, true
}

// GetArgs returns the args left over after a successful parse.
func (fs *CoreFlagSet) GetArgs() []string {
	return fs.postParseArgs
}

// Reset the flag set, and all of the flags defined in the flag set, to their
// initial state. After a successful reset, the flag set may be parsed as if it
// were newly constructed.
func (fs *CoreFlagSet) Reset() error {
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
	// utf8.RuneError is an invalid short name and is ignored.
	//
	// At least one of ShortName and/or LongName is required.
	ShortName rune

	// LongName is the long form name of the flag, which can be provided as a
	// commandline argument with a double-dash -- prefix. An empty string is an
	// invalid long name and is ignored.
	//
	// At least one of ShortName and/or LongName is required.
	LongName string

	// Placeholder is typically used to represent an example value in the help
	// text for the flag. For example, a placeholder of `BAR` might result in
	// help text like
	//
	//      -f, --foo BAR   set the foo parameter
	//
	// Optional.
	Placeholder string

	// Usage is a short help message for the flag, typically printed after the
	// flag name(s) on a single line in the help output. The first `backtick`
	// quoted substring (if any) will be the placeholder, if a placeholder was
	// not explicitly provided.
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
}

// AddFlag adds a flag to the flag set, as specified by the provided config. An
// error is returned if the config is invalid, or if a flag is already defined
// in the flag set with the same short or long name.
//
// This is a fairly low level method. Consumers may prefer type-specific helpers
// like [CoreFlagSet.Bool], [CoreFlagSet.StringVar], etc.
func (fs *CoreFlagSet) AddFlag(cfg CoreFlagConfig) error {
	if fs.isStdAdapter {
		return fmt.Errorf("cannot add flags to standard flag set adapter")
	}

	var (
		hasShort = isValidShortName(cfg.ShortName)
		hasLong  = isValidLongName(cfg.LongName)
	)
	if !hasShort && !hasLong {
		return fmt.Errorf("short name and/or long name is required")
	}

	if cfg.Value == nil {
		return fmt.Errorf("value is required")
	}

	var isBoolFlag bool
	if bf, ok := cfg.Value.(interface{ IsBoolFlag() bool }); ok {
		isBoolFlag = bf.IsBoolFlag()
	}

	if isBoolFlag && !hasLong {
		if defaultTrue, err := strconv.ParseBool(cfg.Value.String()); err == nil && defaultTrue {
			return fmt.Errorf("%s: default true boolean flag requires a long name", string(cfg.ShortName))
		}
	}

	if cfg.Placeholder == "" {
		cfg.Placeholder = getPlaceholderFor(cfg.Value, cfg.Usage)
	}

	f := &coreFlag{
		flagSet:      fs,
		shortName:    cfg.ShortName,
		longName:     cfg.LongName,
		placeholder:  cfg.Placeholder,
		defaultValue: cfg.Value.String(),
		usageText:    cfg.Usage,
		flagValue:    cfg.Value,
		isBoolFlag:   isBoolFlag,
	}

	for _, existing := range fs.flagSlice {
		if isDuplicateCoreFlag(f, existing) {
			return newFlagError(f, ErrDuplicateFlag)
		}
	}

	fs.flagSlice = append(fs.flagSlice, f)

	return nil
}

// BoolVar defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlagSet) BoolVar(pointer *bool, short rune, long string, def bool, usage string) {
	mustDefineFlag(fs, pointer, short, long, def, usage)
}

// Bool defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlagSet) Bool(short rune, long string, def bool, usage string) *bool {
	var value bool
	fs.BoolVar(&value, short, long, def, usage)
	return &value
}

// BoolShort defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlagSet) BoolShort(short rune, def bool, usage string) *bool {
	return fs.Bool(short, "", def, usage)
}

// BoolLong defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlagSet) BoolLong(long string, def bool, usage string) *bool {
	return fs.Bool(0, long, def, usage)
}

// StringVar defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlagSet) StringVar(pointer *string, short rune, long string, def string, usage string) {
	mustDefineFlag(fs, pointer, short, long, def, usage)
}

// String defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlagSet) String(short rune, long string, def string, usage string) *string {
	var value string
	fs.StringVar(&value, short, long, def, usage)
	return &value
}

// StringShort defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlagSet) StringShort(short rune, def string, usage string) *string {
	return fs.String(short, "", def, usage)
}

// StringLong defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlagSet) StringLong(long string, def string, usage string) *string {
	return fs.String(0, long, def, usage)
}

// Float64Var defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlagSet) Float64Var(pointer *float64, short rune, long string, def float64, usage string) {
	mustDefineFlag(fs, pointer, short, long, def, usage)
}

// Float64 defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlagSet) Float64(short rune, long string, def float64, usage string) *float64 {
	var value float64
	fs.Float64Var(&value, short, long, def, usage)
	return &value
}

// Float64Short defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlagSet) Float64Short(short rune, def float64, usage string) *float64 {
	return fs.Float64(short, "", def, usage)
}

// Float64Long defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlagSet) Float64Long(long string, def float64, usage string) *float64 {
	return fs.Float64(0, long, def, usage)
}

// IntVar defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlagSet) IntVar(pointer *int, short rune, long string, def int, usage string) {
	mustDefineFlag(fs, pointer, short, long, def, usage)
}

// Int defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlagSet) Int(short rune, long string, def int, usage string) *int {
	var value int
	fs.IntVar(&value, short, long, def, usage)
	return &value
}

// IntShort defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlagSet) IntShort(short rune, def int, usage string) *int {
	return fs.Int(short, "", def, usage)
}

// IntLong defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlagSet) IntLong(long string, def int, usage string) *int {
	return fs.Int(0, long, def, usage)
}

// UintVar defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlagSet) UintVar(pointer *uint, short rune, long string, def uint, usage string) {
	mustDefineFlag(fs, pointer, short, long, def, usage)
}

// Uint defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlagSet) Uint(short rune, long string, def uint, usage string) *uint {
	var value uint
	fs.UintVar(&value, short, long, def, usage)
	return &value
}

// UintShort defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlagSet) UintShort(short rune, def uint, usage string) *uint {
	return fs.Uint(short, "", def, usage)
}

// UintLong defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlagSet) UintLong(long string, def uint, usage string) *uint {
	return fs.Uint(0, long, def, usage)
}

// Uint64Var defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlagSet) Uint64Var(pointer *uint64, short rune, long string, def uint64, usage string) {
	mustDefineFlag(fs, pointer, short, long, def, usage)
}

// Uint64 defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlagSet) Uint64(short rune, long string, def uint64, usage string) *uint64 {
	var value uint64
	fs.Uint64Var(&value, short, long, def, usage)
	return &value
}

// Uint64Short defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlagSet) Uint64Short(short rune, def uint64, usage string) *uint64 {
	return fs.Uint64(short, "", def, usage)
}

// Uint64Long defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlagSet) Uint64Long(long string, def uint64, usage string) *uint64 {
	return fs.Uint64(0, long, def, usage)
}

// DurationVar defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlagSet) DurationVar(pointer *time.Duration, short rune, long string, def time.Duration, usage string) {
	mustDefineFlag(fs, pointer, short, long, def, usage)
}

// Duration defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlagSet) Duration(short rune, long string, def time.Duration, usage string) *time.Duration {
	var value time.Duration
	fs.DurationVar(&value, short, long, def, usage)
	return &value
}

// DurationShort defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlagSet) DurationShort(short rune, def time.Duration, usage string) *time.Duration {
	return fs.Duration(short, "", def, usage)
}

// DurationLong defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlagSet) DurationLong(long string, def time.Duration, usage string) *time.Duration {
	return fs.Duration(0, long, def, usage)
}

// Func defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlagSet) Func(short rune, long string, fn func(string) error, usage string) {
	stdfs := flag.NewFlagSet("flagset-name", flag.ContinueOnError)
	stdfs.Func("flag-name", "flag-usage", fn)
	value := stdfs.Lookup("flag-name").Value

	if err := fs.AddFlag(CoreFlagConfig{
		LongName:  long,
		ShortName: short,
		Usage:     usage,
		Value:     value,
	}); err != nil {
		panic(err)
	}
}

// FuncShort defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlagSet) FuncShort(short rune, fn func(string) error, usage string) {
	fs.Func(short, "", fn, usage)
}

// FuncLong defines a new flag in the flag set, and panics on any error.
func (fs *CoreFlagSet) FuncLong(long string, fn func(string) error, usage string) {
	fs.Func(0, long, fn, usage)
}

func mustDefineFlag[T FlagValueType](fs *CoreFlagSet, pointer *T, short rune, long string, def T, usage string) {
	value := MakeFlagValue(def, pointer)
	if err := fs.AddFlag(CoreFlagConfig{
		ShortName: short,
		LongName:  long,
		Usage:     usage,
		Value:     value,
	}); err != nil {
		panic(err)
	}
}

// FlagValueType enumerates certain primitive types that can produce a
// [flag.Value] via the [MakeFlagValue] helper.
type FlagValueType interface {
	bool | float64 | int | uint | uint64 | string | time.Duration
}

// MakeFlagValue is a helper function that produces a [flag.Value] for certain
// primitive types. The returned value updates the given pointer when set.
func MakeFlagValue[T FlagValueType](defaultValue T, pointer *T) flag.Value {
	var (
		fs = flag.NewFlagSet("flagset-name", flag.ContinueOnError)
		n  = "flag-name"  // just used to fetch the value from the flag set
		u  = "flag-usage" // not actually used
	)
	switch x := any(defaultValue).(type) {
	case bool:
		fs.BoolVar(any(pointer).(*bool), n, x, u)
	case float64:
		fs.Float64Var(any(pointer).(*float64), n, x, u)
	case int:
		fs.IntVar(any(pointer).(*int), n, x, u)
	case uint:
		fs.UintVar(any(pointer).(*uint), n, x, u)
	case uint64:
		fs.Uint64Var(any(pointer).(*uint64), n, x, u)
	case string:
		fs.StringVar(any(pointer).(*string), n, x, u)
	case time.Duration:
		fs.DurationVar(any(pointer).(*time.Duration), n, x, u)
	default:
		panic(fmt.Errorf("unsupported type %T", defaultValue))
	}

	return fs.Lookup(n).Value
}

//
//
//

type coreFlag struct {
	flagSet      *CoreFlagSet
	shortName    rune
	longName     string
	placeholder  string
	defaultValue string
	usageText    string
	flagValue    flag.Value
	isBoolFlag   bool
	isSet        bool
}

var _ Flag = (*coreFlag)(nil)

func (f *coreFlag) GetFlagSetName() string {
	return f.flagSet.flagSetName
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
	return f.defaultValue
}

func (f *coreFlag) GetUsage() string {
	return f.usageText
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
	if err := f.flagValue.Set(f.defaultValue); err != nil {
		return err
	}
	f.isSet = false
	return nil
}

func isDuplicateCoreFlag(f1, f2 *coreFlag) bool {
	var (
		sameShortName = isValidShortName(f1.shortName) && isValidShortName(f2.shortName) && f1.shortName == f2.shortName
		sameLongName  = isValidLongName(f1.longName) && isValidLongName(f2.longName) && f1.longName == f2.longName
	)
	return sameShortName || sameLongName
}
