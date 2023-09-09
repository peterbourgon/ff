package ff

import (
	"flag"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/peterbourgon/ff/v4/ffval"
)

// CoreFlags is the default implementation of a [Flags]. It's broadly similar to
// a flag.FlagSet, but with additional capabilities inspired by getopt(3).
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
// flag.FlagSet, allowing it to implement the [Flags] interface.
//
// The returned core flag set has slightly different behavior than normal. It's
// a fixed "snapshot" of the provided stdfs, which means it doesn't allow new
// flags to be defined, and won't reflect changes made to the stdfs in the
// future. It treats every flag name as a long name, and treats "-" and
// "--" equivalently when parsing arguments.
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

// NewStructFlags returns a core flag set with the given name, and with flags
// taken from the provided val, which must be a pointer to a struct. See
// [CoreFlags.AddStruct] for more information on how struct tags are parsed. Any
// error results in a panic.
func NewStructFlags(name string, val any) *CoreFlags {
	fs := NewFlags(name)
	if err := fs.AddStruct(val); err != nil {
		panic(err)
	}
	return fs
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
	if utf8.RuneCountInString(name) == 1 {
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
	// commandline argument with a double-dash -- prefix. Long names are trimmed
	// of whitespace via [strings.TrimSpace] prior to evaluation. Empty long
	// names are ignored.
	//
	// At least one of ShortName and/or LongName is required.
	LongName string

	// Usage is a short help message for the flag, typically printed after the
	// flag name(s) on a single line in the help text. For example, a usage
	// string "set the foo parameter" might produce help text as follows.
	//
	//      -f, --foo BAR   set the foo parameter
	//
	// If the usage string contains a `backtick` quoted substring, that
	// substring will be treated as a placeholder, if a placeholder was not
	// otherwise explicitly provided.
	//
	// Recommended.
	Usage string

	// Value is used to parse and store the underlying value for the flag.
	// Package ffval provides helpers and definitions for common value types.
	//
	// Required.
	Value flag.Value

	// Placeholder represents an example value in the help text for the flag,
	// typically printed after the flag name(s). For example, a placeholder of
	// "BAR" might produce help text as follows.
	//
	//      -f, --foo BAR   set the foo parameter
	//
	// Optional.
	Placeholder string

	// NoPlaceholder will force GetPlaceholder to return the empty string. This
	// can be useful for flags that don't need placeholders in their help text,
	// for example boolean flags.
	NoPlaceholder bool

	// NoDefault will force GetDefault to return the empty string. This can be
	// useful for flags whose default values don't need to be communicated in
	// help text.
	NoDefault bool
}

func (cfg CoreFlagConfig) isBoolFlag() bool {
	if bf, ok := cfg.Value.(interface{ IsBoolFlag() bool }); ok {
		return bf.IsBoolFlag()
	}
	return false
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
	if bf, ok := cfg.Value.(interface{ IsBoolFlag() bool }); ok && bf.IsBoolFlag() {
		if b, err := strconv.ParseBool(cfg.Value.String()); err == nil && !b {
			return ""
		}
	}

	// If the flag value provides its own non-empty Placeholder, use that.
	// This has lower priority than the bool flag check, above.
	if ph, ok := cfg.Value.(interface{ GetPlaceholder() string }); ok {
		if p := ph.GetPlaceholder(); p != "" {
			return p
		}
	}

	// Otherwise, use a transformation of the flag value type name.
	var typeName string
	{
		typeName = strings.ToUpper(fmt.Sprintf("%T", cfg.Value))
		typeName = genericTypeNameRegexp.ReplaceAllString(typeName, "$1")
		typeName = strings.TrimSuffix(typeName, "VALUE")
		if lastDot := strings.LastIndex(typeName, "."); lastDot > 0 {
			typeName = typeName[lastDot+1:]
		}
	}
	return typeName
}

var genericTypeNameRegexp = regexp.MustCompile(`[A-Z0-9\_\.\*]+\[(.+)\]`)

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

	cfg.LongName = strings.TrimSpace(cfg.LongName)

	var (
		hasShort    = cfg.ShortName != 0
		hasLong     = cfg.LongName != ""
		validShort  = isValidShortName(cfg.ShortName)
		validLong   = isValidLongName(cfg.LongName)
		isBoolFlag  = cfg.isBoolFlag()
		trueDefault = cfg.Value.String()
	)
	if hasShort && !validShort {
		return nil, fmt.Errorf("-%s: invalid short name", string(cfg.ShortName))
	}
	if hasLong && !validLong {
		return nil, fmt.Errorf("--%s: invalid long name", cfg.LongName)
	}
	if !validShort && !validLong {
		return nil, fmt.Errorf("at least one valid name is required")
	}
	if validShort && validLong && string(cfg.ShortName) == cfg.LongName {
		return nil, fmt.Errorf("-%s, --%s: short name identical to long name", string(cfg.ShortName), cfg.LongName)
	}
	if isBoolFlag && !validLong {
		if b, err := strconv.ParseBool(trueDefault); err == nil && b {
			return nil, fmt.Errorf("-%s: default true boolean flag requires a long name", string(cfg.ShortName))
		}
	}

	f := &coreFlag{
		flagSet:     fs,
		shortName:   cfg.ShortName,
		longName:    cfg.LongName,
		usage:       cfg.Usage,
		flagValue:   cfg.Value,
		trueDefault: trueDefault,
		isBoolFlag:  isBoolFlag,
		isSet:       false,
		placeholder: cfg.getPlaceholder(),
		noDefault:   cfg.NoDefault,
	}

	for _, existing := range fs.flagSlice {
		if isDuplicate(f, existing) {
			return nil, newFlagError(f, fmt.Errorf("%w (%s)", ErrDuplicateFlag, getNameString(existing)))
		}
	}

	fs.flagSlice = append(fs.flagSlice, f)

	return f, nil
}

// AddStruct adds flags to the flag set from the given val, which must be a
// pointer to a struct. Each exported field in that struct with a valid `ff:`
// struct tag corresponds to a unique flag in the flag set. Those fields must be
// a supported [ffval.ValueType] or implement [flag.Value].
//
// The `ff:` struct tag is parsed as a sequence of comma- or pipe-separated
// items. An item is either a key, or a key/value pair. Key/value pairs are
// expressed as either key=value (with =), or key: value (with :). Keys, values,
// and items themselves are trimmed of whitespace before use. Values may be
// 'single quoted' and will be unquoted before use.
//
// The following is a list of valid keys and their expected values.
//
//   - s, short, shortname -- value should be a valid short name
//   - l, long, longname -- value should be a valid long name
//   - u, usage -- value should be a valid usage string
//   - d, def, default -- value should be assignable to the flag
//   - p, placeholder -- value should be a valid placeholder
//   - noplaceholder -- (no value)
//   - nodefault -- (no value)
//
// See the example for more detail.
func (fs *CoreFlags) AddStruct(val any) error {
	outerVal := reflect.ValueOf(val)
	if outerVal.Kind() != reflect.Pointer {
		return fmt.Errorf("value (%T) must be a pointer", val)
	}

	innerVal := outerVal.Elem()
	innerTyp := innerVal.Type()
	if innerVal.Kind() != reflect.Struct {
		return fmt.Errorf("value (%T) must be a struct", innerTyp)
	}

	for i := 0; i < innerVal.NumField(); i++ {
		// Evaluate this struct field.
		var (
			fieldVal  = innerVal.Field(i)
			fieldTyp  = innerTyp.Field(i)
			fieldName = fieldTyp.Name
		)

		// Only care if it has `ff:` tag.
		fftag, ok := fieldTyp.Tag.Lookup("ff")
		if !ok {
			continue
		}

		// Only care if the `ff:` tag has one or more comma-separated items.
		var items []string
		{
			var quoted bool
			items = strings.FieldsFunc(fftag, func(r rune) bool {
				if r == '\'' {
					quoted = !quoted
				}
				return !quoted && (r == ',' || r == '|')
			})
			if len(items) <= 0 {
				continue
			}
		}

		// Parse the items into a flag config.
		var (
			cfg CoreFlagConfig
			def string
		)
		for _, item := range items {
			// Allow the tag string to include padding spaces.
			item = strings.TrimSpace(item)
			if item == "" {
				continue
			}

			var key, val string
			if sep := strings.IndexAny(item, "=:"); sep < 0 {
				key = item
			} else {
				key, val = item[:sep], item[sep+1:]
			}
			{
				key = strings.ToLower(key)
				key = strings.TrimSpace(key)
				if key == "" {
					return fmt.Errorf("%s: %q: no key", fieldName, item)
				}
			}
			{
				val = strings.TrimSpace(val)
				if strings.HasPrefix(val, `'`) && strings.HasSuffix(val, `'`) { // treat 'single-quoted values' same as "double-quoted values"
					val = val[1 : len(val)-1]
				} else if v, err := strconv.Unquote(val); err == nil {
					val = v
				}
			}

			// Parse supported keys.
			switch key {
			case "s", "short", "shortname":
				if n := utf8.RuneCountInString(val); n != 1 {
					return fmt.Errorf("%s: %q: invalid short name", fieldName, item)
				}
				short, _ := utf8.DecodeRuneInString(val)
				cfg.ShortName = short

			case "l", "long", "longname":
				if val == "" {
					return fmt.Errorf("%s: %q: invalid (empty) long name", fieldName, item)
				}
				cfg.LongName = val

			case "u", "usage":
				if val == "" {
					return fmt.Errorf("%s: %q: invalid (empty) usage", fieldName, item)
				}
				cfg.Usage = val

			case "d", "def", "default":
				switch val {
				case "", "-":
					cfg.NoDefault = true
				default:
					def = val
				}

			case "nodefault":
				cfg.NoDefault = true

			case "p", "placeholder":
				switch val {
				case "", "-":
					cfg.NoPlaceholder = true
				default:
					cfg.Placeholder = val
				}

			case "noplaceholder":
				cfg.NoPlaceholder = true

			default:
				return fmt.Errorf("%s: %q: unknown key", fieldName, key)
			}
		}

		// Produce a flag.Value representing the field.
		{
			var (
				fieldValAddr      = fieldVal.Addr()
				fieldValAddrTyp   = fieldValAddr.Type()
				fieldValAddrIface = fieldValAddr.Interface()
				flagValueElemTyp  = reflect.TypeOf((*flag.Value)(nil)).Elem()
			)
			if fieldValAddrTyp.Implements(flagValueElemTyp) {
				// The field implements flag.Value, we can use it directly.
				cfg.Value = fieldValAddrIface.(flag.Value)
			} else {
				// Try to construct a new flag value.
				v, err := ffval.NewValueReflect(fieldValAddrIface, def)
				if err != nil {
					return fmt.Errorf("%s: %w", fieldName, err)
				}
				cfg.Value = v
			}
		}

		// Try to add a flag from the parsed config.
		if _, err := fs.AddFlag(cfg); err != nil {
			return fmt.Errorf("%s: %w", fieldName, err)
		}
	}

	return nil
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
// Bool flags should almost always be default false.
func (fs *CoreFlags) BoolVar(pointer *bool, short rune, long string, def bool, usage string) Flag {
	return fs.Value(short, long, ffval.NewValueDefault(pointer, def), usage)
}

// Bool defines a new flag in the flag set, and panics on any error.
// Bool flags should almost always be default false.
func (fs *CoreFlags) Bool(short rune, long string, def bool, usage string) *bool {
	var value bool
	fs.BoolVar(&value, short, long, def, usage)
	return &value
}

// BoolShort defines a new flag in the flag set, and panics on any error.
// Bool flags should almost always be default false.
func (fs *CoreFlags) BoolShort(short rune, def bool, usage string) *bool {
	return fs.Bool(short, "", def, usage)
}

// BoolLong defines a new flag in the flag set, and panics on any error.
// Bool flags should almost always be default false.
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
// See [CoreFlags.StringListVar] for more details.
func (fs *CoreFlags) StringList(short rune, long string, usage string) *[]string {
	var value []string
	fs.StringListVar(&value, short, long, usage)
	return &value
}

// StringListShort defines a new flag in the flag set, and panics on any error.
// See [CoreFlags.StringListVar] for more details.
func (fs *CoreFlags) StringListShort(short rune, usage string) *[]string {
	return fs.StringList(short, "", usage)
}

// StringListLong defines a new flag in the flag set, and panics on any error.
// See [CoreFlags.StringListVar] for more details.
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
// See [CoreFlags.StringSetVar] for more details.
func (fs *CoreFlags) StringSet(short rune, long string, usage string) *[]string {
	var value []string
	fs.StringSetVar(&value, short, long, usage)
	return &value
}

// StringSetShort defines a new flag in the flag set, and panics on any error.
// See [CoreFlags.StringSetVar] for more details.
func (fs *CoreFlags) StringSetShort(short rune, usage string) *[]string {
	return fs.StringSet(short, "", usage)
}

// StringSetLong defines a new flag in the flag set, and panics on any error.
// See [CoreFlags.StringSetVar] for more details.
func (fs *CoreFlags) StringSetLong(long string, usage string) *[]string {
	return fs.StringSet(0, long, usage)
}

// StringEnumVar defines a new enum in the flag set, and panics on any error.
// The default is the first valid value. At least one valid value is required.
func (fs *CoreFlags) StringEnumVar(pointer *string, short rune, long string, usage string, valid ...string) Flag {
	return fs.Value(short, long, ffval.NewEnum(pointer, valid...), usage)
}

// StringEnum defines a new enum in the flag set, and panics on any error.
// The default is the first valid value. At least one valid value is required.
func (fs *CoreFlags) StringEnum(short rune, long string, usage string, valid ...string) *string {
	var value string
	fs.StringEnumVar(&value, short, long, usage, valid...)
	return &value
}

// StringEnumShort defines a new enum in the flag set, and panics on any error.
// The default is the first valid value. At least one valid value is required.
func (fs *CoreFlags) StringEnumShort(short rune, usage string, valid ...string) *string {
	return fs.StringEnum(short, "", usage, valid...)
}

// StringEnumLong defines a new enum in the flag set, and panics on any error.
// The default is the first valid value. At least one valid value is required.
func (fs *CoreFlags) StringEnumLong(long string, usage string, valid ...string) *string {
	return fs.StringEnum(0, long, usage, valid...)
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
	usage       string
	flagValue   flag.Value
	trueDefault string // actual default, for e.g. Reset
	isBoolFlag  bool
	isSet       bool
	placeholder string
	noDefault   bool // in help text
}

var _ Flag = (*coreFlag)(nil)
var _ Resetter = (*coreFlag)(nil)

func (f *coreFlag) GetFlags() Flags {
	return f.flagSet
}

func (f *coreFlag) GetShortName() (rune, bool) {
	return f.shortName, isValidShortName(f.shortName)
}

func (f *coreFlag) GetLongName() (string, bool) {
	return f.longName, isValidLongName(f.longName)
}

func (f *coreFlag) GetUsage() string {
	return f.usage
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
	} else {
		if err := f.flagValue.Set(f.trueDefault); err != nil {
			return err
		}
	}

	f.isSet = false
	return nil
}

func (f *coreFlag) GetPlaceholder() string {
	return f.placeholder
}

func (f *coreFlag) GetDefault() string {
	if f.noDefault {
		return ""
	}
	return f.trueDefault
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
