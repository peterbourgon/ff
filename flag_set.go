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

// FlagSet is a standard implementation of [Flags]. It's broadly similar to a
// flag.FlagSet, but with additional capabilities inspired by getopt(3).
type FlagSet struct {
	name          string
	flags         []*coreFlag
	isParsed      bool
	postParseArgs []string
	isStdAdapter  bool // stdlib package flag behavior: treat -foo the same as --foo
	parent        *FlagSet
}

var _ Flags = (*FlagSet)(nil)
var _ Resetter = (*FlagSet)(nil)

// NewFlagSet returns a new flag set with the given name.
func NewFlagSet(name string) *FlagSet {
	return &FlagSet{
		name:          name,
		flags:         []*coreFlag{},
		isParsed:      false,
		postParseArgs: []string{},
		isStdAdapter:  false,
		parent:        nil,
	}
}

// NewFlagSetFrom is a helper method that calls [NewFlagSet] with name, and then
// [FlagSet.AddStruct] with val, which must be a pointer to a struct. Any error
// results in a panic.
//
// As a special case, val may also be a pointer to a flag.FlagSet. In this case,
// the returned ff.FlagSet behaves differently than normal. It acts as a fixed
// "snapshot" of the flag.FlagSet, and so doesn't allow new flags to be added.
// To approximate the behavior of the standard library, every flag.FlagSet flag
// name is treated as a long name, and parsing treats single-hyphen and
// double-hyphen flag arguments identically: -abc is parsed as --abc rather than
// -a -b -c. The flag.FlagSet error handling strategy is (effectively) forced to
// ContinueOnError. The usage function is ignored, and usage is never printed as
// a side effect of parsing.
func NewFlagSetFrom(name string, val any) *FlagSet {
	if stdfs, ok := val.(*flag.FlagSet); ok {
		if name == "" {
			name = stdfs.Name()
		}
		corefs := NewFlagSet(name)
		stdfs.VisitAll(func(f *flag.Flag) {
			if _, err := corefs.AddFlag(FlagConfig{
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

	fs := NewFlagSet(name)
	if err := fs.AddStruct(val); err != nil {
		panic(err)
	}
	return fs
}

// SetParent assigns a parent flag set to this one, making all parent flags
// available, recursively, to the receiver. For example, Parse will match
// against any parent flag, WalkFlags will traverse all parent flags, etc.
//
// This method returns its receiver to allow for builder-style initialization.
func (fs *FlagSet) SetParent(parent *FlagSet) *FlagSet {
	fs.parent = parent
	return fs
}

// GetName returns the name of the flag set provided during construction.
func (fs *FlagSet) GetName() string {
	return fs.name
}

// Parse the provided args against the flag set, assigning flag values as
// appropriate. Args are matched to flags defined in this flag set, and, if a
// parent is set, all parent flag sets, recursively. If a specified flag can't
// be found, parse fails with [ErrUnknownFlag]. After a successful parse,
// subsequent calls to parse fail with [ErrAlreadyParsed], until and unless the
// flag set is reset.
func (fs *FlagSet) Parse(args []string) error {
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

func (fs *FlagSet) parseArgs(args []string) (err error) {
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
func (fs *FlagSet) findFlag(short rune, long string) *coreFlag {
	var (
		haveShort = isValidShortName(short)
		haveLong  = isValidLongName(long)
	)
	for cursor := fs; cursor != nil; cursor = cursor.parent {
		for _, candidate := range cursor.flags {
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

func (fs *FlagSet) findShortFlag(short rune) *coreFlag {
	return fs.findFlag(short, "")
}

func (fs *FlagSet) findLongFlag(long string) *coreFlag {
	return fs.findFlag(0, long)
}

func (fs *FlagSet) parseShortFlag(arg string, args []string) ([]string, error) {
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

func (fs *FlagSet) parseLongFlag(arg string, args []string) ([]string, error) {
	name, value, eqFound := strings.Cut(arg, "=")
	name = strings.TrimPrefix(name, "--")

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

	if eqFound && f.isBoolFlag && value == "" {
		value = "true" // `--debug=` amounts to `--debug=true`
	}

	if value == "" && !eqFound {
		switch {
		case f.isBoolFlag:
			value = "true" // `--foo` defaults to true
			if len(args) > 0 {
				if _, err := strconv.ParseBool(args[0]); err == nil {
					value = args[0] // `--foo false` should also work
					args = args[1:]
				}
			}
		case len(args) > 0:
			value, args = args[0], args[1:]
		default:
			return nil, fmt.Errorf("missing value")
		}
	}

	if err := f.flagValue.Set(value); err != nil {
		return nil, newFlagError(f, fmt.Errorf("set %q: %w", value, err))
	}
	f.isSet = true

	return args, nil
}

// IsParsed returns true if the flag set has been successfully parsed.
func (fs *FlagSet) IsParsed() bool {
	return fs.isParsed
}

// WalkFlags calls fn for every flag known to the flag set. This includes all
// parent flags, if a parent has been set.
func (fs *FlagSet) WalkFlags(fn func(Flag) error) error {
	for cursor := fs; cursor != nil; cursor = cursor.parent {
		for _, f := range cursor.flags {
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
func (fs *FlagSet) GetFlag(name string) (Flag, bool) {
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
func (fs *FlagSet) GetArgs() []string {
	return fs.postParseArgs
}

// Reset the flag set, and all of the flags defined in the flag set, to their
// initial state. After a successful reset, the flag set may be parsed as if it
// were newly constructed.
func (fs *FlagSet) Reset() error {
	for _, f := range fs.flags {
		if err := f.Reset(); err != nil {
			return newFlagError(f, err)
		}
	}

	fs.postParseArgs = fs.postParseArgs[:0]
	fs.isParsed = false

	return nil
}

// FlagConfig collects the required config for a flag in a flag set.
type FlagConfig struct {
	// ShortName is the short form name of the flag, which can be provided as a
	// commandline argument with a single dash - prefix. A rune value of 0 or
	// utf8.RuneError is considered an invalid short name and is ignored.
	//
	// At least one of ShortName and/or LongName is required.
	ShortName rune

	// LongName is the long form name of the flag, which can be provided as a
	// commandline argument with a double-dash -- prefix. Long names must be
	// non-empty, and cannot contain whitespace, control characters, single or
	// double quotes, backticks, or backslashes.
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
	// Optional. If not provided, a default based on the value type is used.
	Placeholder string

	// NoPlaceholder will force GetPlaceholder to return the empty string. This
	// can be useful for flags that don't need placeholders in their help text,
	// for example boolean flags.
	NoPlaceholder bool

	// NoDefault will force GetDefault to return the empty string. This can be
	// useful for flags whose default values don't need to be communicated in
	// help text. Note this does not affect the actual default value of the
	// flag.
	NoDefault bool
}

func (cfg FlagConfig) isBoolFlag() bool {
	if bf, ok := cfg.Value.(interface{ IsBoolFlag() bool }); ok {
		return bf.IsBoolFlag()
	}
	return false
}

func (cfg FlagConfig) getPlaceholder() string {
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

func (cfg FlagConfig) getHelpDefault() string {
	// If a default is explicitly refused, use an empty string.
	if cfg.NoDefault {
		return ""
	}

	// Bool flags with default value false should have empty defaults.
	if bf, ok := cfg.Value.(interface{ IsBoolFlag() bool }); ok && bf.IsBoolFlag() {
		if b, err := strconv.ParseBool(cfg.Value.String()); err == nil && !b {
			return ""
		}
	}

	// Otherwise, use the flag value.
	return cfg.Value.String()
}

var genericTypeNameRegexp = regexp.MustCompile(`[A-Z0-9\_\.\*]+\[(.+)\]`)

// AddFlag adds a flag to the flag set, as specified by the provided config. An
// error is returned if the config is invalid, or if a flag is already defined
// in the flag set with the same short or long name.
//
// This is a fairly low level method. Consumers may prefer type-specific helpers
// like [FlagSet.Bool], [FlagSet.StringVar], etc.
func (fs *FlagSet) AddFlag(cfg FlagConfig) (Flag, error) {
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
		return nil, fmt.Errorf("-%s, --%s: same short and long name", string(cfg.ShortName), cfg.LongName)
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
		helpDefault: cfg.getHelpDefault(),
	}

	for _, existing := range fs.flags {
		if isDuplicate(f, existing) {
			return nil, newFlagError(f, fmt.Errorf("%w (%s)", ErrDuplicateFlag, getNameString(existing)))
		}
	}

	fs.flags = append(fs.flags, f)

	return f, nil
}

// AddStruct adds flags to the flag set from the given val, which must be a
// pointer to a struct. Each exported field in that struct with a valid `ff:`
// struct tag corresponds to a unique flag in the flag set. Those fields must be
// a supported [ffval.ValueType] or implement [flag.Value].
//
// The `ff:` struct tag is a sequence of comma- or pipe-delimited items. An item
// is either empty (and ignored), a key, or a key/value pair. Key/value pairs
// are expressed as either key=value (with =) or key:value (with :). Items,
// keys, and values are trimmed of whitespace before use. Values may be 'single
// quoted' and will be unquoted before use.
//
// The following is a list of valid keys and their expected values. Any invalid
// item, key, or value in the `ff:` struct tag will result in an error.
//
//   - s, short, shortname -- value must be a single valid rune
//   - l, long, longname -- value must be a valid long name
//   - u, usage -- value must be a non-empty string
//   - d, def, default -- value must be a non-empty and assignable string
//   - p, placeholder -- value must be a non-empty string
//   - noplaceholder -- no value
//   - nodefault -- no value
//
// See the example for more detail.
func (fs *FlagSet) AddStruct(val any) error {
	outerVal := reflect.ValueOf(val)
	if outerVal.Kind() != reflect.Pointer {
		return fmt.Errorf("value (%T) must be a pointer", val)
	}

	innerVal := outerVal.Elem()
	innerTyp := innerVal.Type()
	if innerVal.Kind() != reflect.Struct {
		return fmt.Errorf("value (%T) must be a struct", innerTyp)
	}

	// We'll collect flag configs in one pass, and add the flags afterwards.
	var flagConfigs []FlagConfig

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
			cfg FlagConfig
			def string
		)
		for _, item := range items {
			// Allow spaces for padding.
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
					return fmt.Errorf("%s: %s: invalid (empty) long name", fieldName, item)
				}
				cfg.LongName = val

			case "u", "usage":
				if val == "" {
					return fmt.Errorf("%s: %s: invalid (empty) usage", fieldName, item)
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
				if val != "" {
					return fmt.Errorf("%s: %s: nodefault should not have a value", fieldName, item)
				}
				cfg.NoDefault = true

			case "p", "placeholder":
				switch val {
				case "", "-":
					cfg.NoPlaceholder = true
				default:
					cfg.Placeholder = val
				}

			case "noplaceholder":
				if val != "" {
					return fmt.Errorf("%s: %s: noplaceholder should not have a value", fieldName, item)
				}
				cfg.NoPlaceholder = true

			default:
				return fmt.Errorf("%s: %s: unknown key", fieldName, key)
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

		// Save the config to add later, after the struct is fully parsed.
		flagConfigs = append(flagConfigs, cfg)
	}

	// Add the collected flags.
	for _, cfg := range flagConfigs {
		if _, err := fs.AddFlag(cfg); err != nil {
			return err
		}
	}

	return nil
}

// Value defines a new flag in the flag set, and panics on any error.
func (fs *FlagSet) Value(short rune, long string, value flag.Value, usage string) Flag {
	f, err := fs.AddFlag(FlagConfig{
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
func (fs *FlagSet) ValueShort(short rune, value flag.Value, usage string) Flag {
	return fs.Value(short, "", value, usage)
}

// ValueLong defines a new flag in the flag set, and panics on any error.
func (fs *FlagSet) ValueLong(long string, value flag.Value, usage string) Flag {
	return fs.Value(0, long, value, usage)
}

// BoolVar defines a new default false bool flag in the flag set, and panics on
// any error.
func (fs *FlagSet) BoolVar(pointer *bool, short rune, long string, usage string) Flag {
	return fs.BoolVarDefault(pointer, short, long, false, usage)
}

// BoolVarDefault defines a new bool flag in the flag set, and panics on any
// error. Bool flags should almost always be default false; prefer BoolVar to
// BoolVarDefault.
func (fs *FlagSet) BoolVarDefault(pointer *bool, short rune, long string, def bool, usage string) Flag {
	return fs.Value(short, long, ffval.NewValueDefault(pointer, def), usage)
}

// Bool defines a new default false bool flag in the flag set, and panics on any
// error.
func (fs *FlagSet) Bool(short rune, long string, usage string) *bool {
	return fs.BoolDefault(short, long, false, usage)
}

// BoolDefault defines a new bool flag in the flag set, and panics on any error.
// Bool flags should almost always be default false; prefer Bool to BoolDefault.
func (fs *FlagSet) BoolDefault(short rune, long string, def bool, usage string) *bool {
	var value bool
	fs.BoolVarDefault(&value, short, long, def, usage)
	return &value
}

// BoolShort defines a new default false bool flag in the flag set, and panics
// on any error.
func (fs *FlagSet) BoolShort(short rune, usage string) *bool {
	return fs.Bool(short, "", usage)
}

// BoolShortDefault defines a new bool flag in the flag set, and panics on any
// error. Bool flags should almost always be default false; prefer BoolShort to
// BoolShortDefault.
func (fs *FlagSet) BoolShortDefault(short rune, def bool, usage string) *bool {
	return fs.BoolDefault(short, "", def, usage)
}

// BoolLong defines a new default false bool flag in the flag set, and panics on
// any error.
func (fs *FlagSet) BoolLong(long string, usage string) *bool {
	return fs.Bool(0, long, usage)
}

// BoolLongDefault defines a new bool flag in the flag set, and panics on any
// error. Bool flags should almost always be default false; prefer BoolLong to
// BoolLongDefault.
func (fs *FlagSet) BoolLongDefault(long string, def bool, usage string) *bool {
	return fs.BoolDefault(0, long, def, usage)
}

// BoolConfig defines a new flag in the flag set, and panics on any error.
// The value field of the provided config is overwritten.
func (fs *FlagSet) BoolConfig(cfg FlagConfig) *bool {
	var value bool
	cfg.Value = ffval.NewValue(&value)
	if _, err := fs.AddFlag(cfg); err != nil {
		panic(err)
	}
	return &value
}

// StringVar defines a new flag in the flag set, and panics on any error.
func (fs *FlagSet) StringVar(pointer *string, short rune, long string, def string, usage string) Flag {
	return fs.Value(short, long, ffval.NewValueDefault(pointer, def), usage)
}

// String defines a new flag in the flag set, and panics on any error.
func (fs *FlagSet) String(short rune, long string, def string, usage string) *string {
	var value string
	fs.StringVar(&value, short, long, def, usage)
	return &value
}

// StringShort defines a new flag in the flag set, and panics on any error.
func (fs *FlagSet) StringShort(short rune, def string, usage string) *string {
	return fs.String(short, "", def, usage)
}

// StringLong defines a new flag in the flag set, and panics on any error.
func (fs *FlagSet) StringLong(long string, def string, usage string) *string {
	return fs.String(0, long, def, usage)
}

// StringConfig defines a new flag in the flag set, and panics on any error.
// The value field of the provided config is overwritten.
func (fs *FlagSet) StringConfig(cfg FlagConfig, def string) *string {
	var value string
	cfg.Value = ffval.NewValueDefault(&value, def)
	if _, err := fs.AddFlag(cfg); err != nil {
		panic(err)
	}
	return &value
}

// StringListVar defines a new flag in the flag set, and panics on any error.
//
// The flag represents a list of strings, where each call to Set adds a new
// value to the list. Duplicate values are permitted.
func (fs *FlagSet) StringListVar(pointer *[]string, short rune, long string, usage string) Flag {
	return fs.Value(short, long, ffval.NewList(pointer), usage)
}

// StringList defines a new flag in the flag set, and panics on any error.
// See [FlagSet.StringListVar] for more details.
func (fs *FlagSet) StringList(short rune, long string, usage string) *[]string {
	var value []string
	fs.StringListVar(&value, short, long, usage)
	return &value
}

// StringListShort defines a new flag in the flag set, and panics on any error.
// See [FlagSet.StringListVar] for more details.
func (fs *FlagSet) StringListShort(short rune, usage string) *[]string {
	return fs.StringList(short, "", usage)
}

// StringListLong defines a new flag in the flag set, and panics on any error.
// See [FlagSet.StringListVar] for more details.
func (fs *FlagSet) StringListLong(long string, usage string) *[]string {
	return fs.StringList(0, long, usage)
}

// StringSetVar defines a new flag in the flag set, and panics on any error.
//
// The flag represents a unique list of strings, where each call to Set adds a
// new value to the list. Duplicate values are silently dropped.
func (fs *FlagSet) StringSetVar(pointer *[]string, short rune, long string, usage string) Flag {
	return fs.Value(short, long, ffval.NewUniqueList(pointer), usage)
}

// StringSet defines a new flag in the flag set, and panics on any error.
// See [FlagSet.StringSetVar] for more details.
func (fs *FlagSet) StringSet(short rune, long string, usage string) *[]string {
	var value []string
	fs.StringSetVar(&value, short, long, usage)
	return &value
}

// StringSetShort defines a new flag in the flag set, and panics on any error.
// See [FlagSet.StringSetVar] for more details.
func (fs *FlagSet) StringSetShort(short rune, usage string) *[]string {
	return fs.StringSet(short, "", usage)
}

// StringSetLong defines a new flag in the flag set, and panics on any error.
// See [FlagSet.StringSetVar] for more details.
func (fs *FlagSet) StringSetLong(long string, usage string) *[]string {
	return fs.StringSet(0, long, usage)
}

// StringEnumVar defines a new enum in the flag set, and panics on any error.
// The default is the first valid value. At least one valid value is required.
func (fs *FlagSet) StringEnumVar(pointer *string, short rune, long string, usage string, valid ...string) Flag {
	return fs.Value(short, long, ffval.NewEnum(pointer, valid...), usage)
}

// StringEnum defines a new enum in the flag set, and panics on any error.
// The default is the first valid value. At least one valid value is required.
func (fs *FlagSet) StringEnum(short rune, long string, usage string, valid ...string) *string {
	var value string
	fs.StringEnumVar(&value, short, long, usage, valid...)
	return &value
}

// StringEnumShort defines a new enum in the flag set, and panics on any error.
// The default is the first valid value. At least one valid value is required.
func (fs *FlagSet) StringEnumShort(short rune, usage string, valid ...string) *string {
	return fs.StringEnum(short, "", usage, valid...)
}

// StringEnumLong defines a new enum in the flag set, and panics on any error.
// The default is the first valid value. At least one valid value is required.
func (fs *FlagSet) StringEnumLong(long string, usage string, valid ...string) *string {
	return fs.StringEnum(0, long, usage, valid...)
}

// Float64Var defines a new flag in the flag set, and panics on any error.
func (fs *FlagSet) Float64Var(pointer *float64, short rune, long string, def float64, usage string) Flag {
	return fs.Value(short, long, ffval.NewValueDefault(pointer, def), usage)
}

// Float64 defines a new flag in the flag set, and panics on any error.
func (fs *FlagSet) Float64(short rune, long string, def float64, usage string) *float64 {
	var value float64
	fs.Float64Var(&value, short, long, def, usage)
	return &value
}

// Float64Short defines a new flag in the flag set, and panics on any error.
func (fs *FlagSet) Float64Short(short rune, def float64, usage string) *float64 {
	return fs.Float64(short, "", def, usage)
}

// Float64Long defines a new flag in the flag set, and panics on any error.
func (fs *FlagSet) Float64Long(long string, def float64, usage string) *float64 {
	return fs.Float64(0, long, def, usage)
}

// Float64Config defines a new flag in the flag set, and panics on any error.
// The value field of the provided config is overwritten.
func (fs *FlagSet) Float64Config(cfg FlagConfig, def float64) *float64 {
	var value float64
	cfg.Value = ffval.NewValueDefault(&value, def)
	if _, err := fs.AddFlag(cfg); err != nil {
		panic(err)
	}
	return &value
}

// IntVar defines a new flag in the flag set, and panics on any error.
func (fs *FlagSet) IntVar(pointer *int, short rune, long string, def int, usage string) Flag {
	return fs.Value(short, long, ffval.NewValueDefault(pointer, def), usage)
}

// Int defines a new flag in the flag set, and panics on any error.
func (fs *FlagSet) Int(short rune, long string, def int, usage string) *int {
	var value int
	fs.IntVar(&value, short, long, def, usage)
	return &value
}

// IntShort defines a new flag in the flag set, and panics on any error.
func (fs *FlagSet) IntShort(short rune, def int, usage string) *int {
	return fs.Int(short, "", def, usage)
}

// IntLong defines a new flag in the flag set, and panics on any error.
func (fs *FlagSet) IntLong(long string, def int, usage string) *int {
	return fs.Int(0, long, def, usage)
}

// IntConfig defines a new flag in the flag set, and panics on any error.
// The value field of the provided config is overwritten.
func (fs *FlagSet) IntConfig(cfg FlagConfig, def int) *int {
	var value int
	cfg.Value = ffval.NewValueDefault(&value, def)
	if _, err := fs.AddFlag(cfg); err != nil {
		panic(err)
	}
	return &value
}

// Int64Var defines a new flag in the flag set, and panics on any error.
func (fs *FlagSet) Int64Var(pointer *int64, short rune, long string, def int64, usage string) Flag {
	return fs.Value(short, long, ffval.NewValueDefault(pointer, def), usage)
}

// Int64 defines a new flag in the flag set, and panics on any error.
func (fs *FlagSet) Int64(short rune, long string, def int64, usage string) *int64 {
	var value int64
	fs.Int64Var(&value, short, long, def, usage)
	return &value
}

// Int64Short defines a new flag in the flag set, and panics on any error.
func (fs *FlagSet) Int64Short(short rune, def int64, usage string) *int64 {
	return fs.Int64(short, "", def, usage)
}

// Int64Long defines a new flag in the flag set, and panics on any error.
func (fs *FlagSet) Int64Long(long string, def int64, usage string) *int64 {
	return fs.Int64(0, long, def, usage)
}

// Int64Config defines a new flag in the flag set, and panics on any error.
// The value field of the provided config is overwritten.
func (fs *FlagSet) Int64Config(cfg FlagConfig, def int64) *int64 {
	var value int64
	cfg.Value = ffval.NewValueDefault(&value, def)
	if _, err := fs.AddFlag(cfg); err != nil {
		panic(err)
	}
	return &value
}

// UintVar defines a new flag in the flag set, and panics on any error.
func (fs *FlagSet) UintVar(pointer *uint, short rune, long string, def uint, usage string) Flag {
	return fs.Value(short, long, ffval.NewValueDefault(pointer, def), usage)
}

// Uint defines a new flag in the flag set, and panics on any error.
func (fs *FlagSet) Uint(short rune, long string, def uint, usage string) *uint {
	var value uint
	fs.UintVar(&value, short, long, def, usage)
	return &value
}

// UintShort defines a new flag in the flag set, and panics on any error.
func (fs *FlagSet) UintShort(short rune, def uint, usage string) *uint {
	return fs.Uint(short, "", def, usage)
}

// UintLong defines a new flag in the flag set, and panics on any error.
func (fs *FlagSet) UintLong(long string, def uint, usage string) *uint {
	return fs.Uint(0, long, def, usage)
}

// UintConfig defines a new flag in the flag set, and panics on any error.
// The value field of the provided config is overwritten.
func (fs *FlagSet) UintConfig(cfg FlagConfig, def uint) *uint {
	var value uint
	cfg.Value = ffval.NewValueDefault(&value, def)
	if _, err := fs.AddFlag(cfg); err != nil {
		panic(err)
	}
	return &value
}

// Uint64Var defines a new flag in the flag set, and panics on any error.
func (fs *FlagSet) Uint64Var(pointer *uint64, short rune, long string, def uint64, usage string) Flag {
	return fs.Value(short, long, ffval.NewValueDefault(pointer, def), usage)
}

// Uint64 defines a new flag in the flag set, and panics on any error.
func (fs *FlagSet) Uint64(short rune, long string, def uint64, usage string) *uint64 {
	var value uint64
	fs.Uint64Var(&value, short, long, def, usage)
	return &value
}

// Uint64Short defines a new flag in the flag set, and panics on any error.
func (fs *FlagSet) Uint64Short(short rune, def uint64, usage string) *uint64 {
	return fs.Uint64(short, "", def, usage)
}

// Uint64Long defines a new flag in the flag set, and panics on any error.
func (fs *FlagSet) Uint64Long(long string, def uint64, usage string) *uint64 {
	return fs.Uint64(0, long, def, usage)
}

// Uint64Config defines a new flag in the flag set, and panics on any error.
// The value field of the provided config is overwritten.
func (fs *FlagSet) Uint64Config(cfg FlagConfig, def uint64) *uint64 {
	var value uint64
	cfg.Value = ffval.NewValueDefault(&value, def)
	if _, err := fs.AddFlag(cfg); err != nil {
		panic(err)
	}
	return &value
}

// DurationVar defines a new flag in the flag set, and panics on any error.
func (fs *FlagSet) DurationVar(pointer *time.Duration, short rune, long string, def time.Duration, usage string) Flag {
	return fs.Value(short, long, ffval.NewValueDefault(pointer, def), usage)
}

// Duration defines a new flag in the flag set, and panics on any error.
func (fs *FlagSet) Duration(short rune, long string, def time.Duration, usage string) *time.Duration {
	var value time.Duration
	fs.DurationVar(&value, short, long, def, usage)
	return &value
}

// DurationShort defines a new flag in the flag set, and panics on any error.
func (fs *FlagSet) DurationShort(short rune, def time.Duration, usage string) *time.Duration {
	return fs.Duration(short, "", def, usage)
}

// DurationLong defines a new flag in the flag set, and panics on any error.
func (fs *FlagSet) DurationLong(long string, def time.Duration, usage string) *time.Duration {
	return fs.Duration(0, long, def, usage)
}

// DurationConfig defines a new flag in the flag set, and panics on any error.
// The value field of the provided config is overwritten.
func (fs *FlagSet) DurationConfig(cfg FlagConfig, def time.Duration) *time.Duration {
	var value time.Duration
	cfg.Value = ffval.NewValueDefault(&value, def)
	if _, err := fs.AddFlag(cfg); err != nil {
		panic(err)
	}
	return &value
}

// Func defines a new flag in the flag set, and panics on any error.
func (fs *FlagSet) Func(short rune, long string, fn func(string) error, usage string) {
	stdfs := flag.NewFlagSet("flagset-name", flag.ContinueOnError)
	stdfs.Func("flag-name", "flag-usage", fn)
	value := stdfs.Lookup("flag-name").Value
	fs.Value(short, long, value, usage)
}

// FuncShort defines a new flag in the flag set, and panics on any error.
func (fs *FlagSet) FuncShort(short rune, fn func(string) error, usage string) {
	fs.Func(short, "", fn, usage)
}

// FuncLong defines a new flag in the flag set, and panics on any error.
func (fs *FlagSet) FuncLong(long string, fn func(string) error, usage string) {
	fs.Func(0, long, fn, usage)
}

//
//
//

type coreFlag struct {
	flagSet     *FlagSet
	shortName   rune
	longName    string
	usage       string
	flagValue   flag.Value
	trueDefault string // actual default, for e.g. Reset
	isBoolFlag  bool
	isSet       bool
	placeholder string
	helpDefault string // string used in help text
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
	return f.helpDefault
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
