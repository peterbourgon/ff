package ff

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	iofs "io/fs"
	"os"
	"strings"
)

// FlagSetAny must be either a [Flags] interface, or a concrete [*flag.FlagSet].
// Any other value will produce a runtime error.
//
// The intent is to make the signature of functions like [Parse] more intuitive.
type FlagSetAny any

// Parse the flag set with the provided args. [Option] values can be used to
// influence parse behavior. For example, options exist to read flags from
// environment variables, config files, etc.
func Parse(fs FlagSetAny, args []string, options ...Option) error {
	switch reified := fs.(type) {
	case Flags:
		return parse(reified, args, options...)
	case *flag.FlagSet:
		return parse(NewFlagSetFrom(reified.Name(), reified), args, options...)
	default:
		return fmt.Errorf("unsupported flag set %T", fs)
	}
}

func parse(fs Flags, args []string, options ...Option) error {
	// The parse context manages options.
	var pc ParseContext
	for _, option := range options {
		option(&pc)
	}

	// Index valid flags by env var key, to support .env config files (below).
	env2flags := map[string][]Flag{}
	{
		// First, collect flags by env var key.
		if err := fs.WalkFlags(func(f Flag) error {
			for _, key := range getEnvVarKeys(f, pc) {
				env2flags[key] = append(env2flags[key], f)
			}
			return nil
		}); err != nil {
			return err
		}

		// If env var support is enabled, to prevent ambiguity, we need to
		// ensure that no two flags share the same env var key.
		//
		// Arguably this check should also be performed if we're using an .env
		// config file parser, but we have no way of knowing that, and special
		// casing *our* ffenv parser isn't a good solution. But this is fine: if
		// a config file specifies a key that maps to more than 1 flag, we can
		// return an error at that point.
		if pc.envVarEnabled {
			for key, flags := range env2flags {
				if len(flags) > 1 {
					return fmt.Errorf("%s: %w (%s) (%s)", getNameString(flags[0]), ErrDuplicateFlag, getNameString(flags[1]), key)
				}
			}
		}
	}

	// After each stage of parsing, record the flags that have been provided.
	// Subsequent lower-priority stages can't set these already-provided flags.
	var provided flagSetSlice
	markProvided := func() {
		fs.WalkFlags(func(f Flag) error {
			if f.IsSet() {
				provided.add(f)
			}
			return nil
		})
	}

	// First priority: the commandline, i.e. the user.
	{
		if err := fs.Parse(args); err != nil {
			return fmt.Errorf("parse args: %w", err)
		}

		markProvided()
	}

	// Second priority: the environment, i.e. the session.
	{
		if pc.envVarEnabled {
			if err := fs.WalkFlags(func(f Flag) error {
				// If the flag has already been set, we can't do anything.
				if provided.has(f) {
					return nil
				}

				// Look in the environment for each of the flag's keys.
				for _, key := range getEnvVarKeys(f, pc) {
					// Look up the value from the environment.
					val, ok := os.LookupEnv(key)
					if !ok {
						continue
					}

					// The value may need to be split.
					vals := []string{val}
					if pc.envVarSplit != "" {
						vals = splitEscape(val, pc.envVarSplit)
					}

					// Set the flag to the value(s).
					for _, v := range vals {
						if err := f.SetValue(v); err != nil {
							return fmt.Errorf("%s=%q: %w", key, val, err)
						}
					}
				}

				return nil
			}); err != nil {
				return fmt.Errorf("parse environment: %w", err)
			}
		}

		markProvided()
	}

	// Third priority: the config file, i.e. the host.
	{
		// First, prefer an explicit filename string.
		var configFile string
		if pc.configFileName != "" {
			configFile = pc.configFileName
		}

		// Next, check the flag name.
		if configFile == "" && pc.configFlagName != "" {
			if f, ok := fs.GetFlag(pc.configFlagName); ok {
				configFile = f.GetValue()
			}
		}

		// If they didn't provide an open func, set the default.
		if pc.configOpenFunc == nil {
			pc.configOpenFunc = func(s string) (iofs.File, error) {
				return os.Open(s)
			}
		}

		// Config files require both a filename and a parser.
		var (
			haveConfigFile  = configFile != ""
			haveParser      = pc.configParseFunc != nil
			parseConfigFile = haveConfigFile && haveParser
		)
		if parseConfigFile {
			configFile, err := pc.configOpenFunc(configFile)
			switch {
			case err == nil:
				defer configFile.Close()
				if err := pc.configParseFunc(configFile, func(name, value string) error {
					var (
						setFlag, fromSet  = fs.GetFlag(name)
						envFlags, fromEnv = env2flags[name]
						target            Flag
					)
					switch {
					case !pc.configKeyIgnoreFlagNames && fromSet: // found in the flag set
						target = setFlag
					case !pc.configKeyIgnoreEnvVars && fromEnv: // found in the environment
						switch len(envFlags) {
						case 0:
							panic(fmt.Errorf("invalid env flag state for %s", name))
						case 1:
							target = envFlags[0]
						default:
							return fmt.Errorf("%s: %w", name, ErrAmbiguousFlag)
						}
					case pc.configIgnoreUndefinedFlags: // not found, but that's OK
						return nil
					case !pc.configIgnoreUndefinedFlags: // not found, and that's not OK
						return fmt.Errorf("%s: %w", name, ErrUnknownFlag)
					default:
						panic(fmt.Errorf("unexpected unreachable case for %s", name))
					}

					// If the flag was already provided by commandline args or
					// env vars, then don't set it again. But be sure to allow
					// config files to specify the same flag multiple times.
					if provided.has(target) {
						return nil
					}

					if err := target.SetValue(value); err != nil {
						return fmt.Errorf("%s: %w", name, err)
					}

					return nil
				}); err != nil {
					return fmt.Errorf("parse config file: %w", err)
				}

			case errors.Is(err, iofs.ErrNotExist) && pc.configAllowMissingFile:
				// no problem

			default:
				return err
			}
		}

		markProvided()
	}

	return nil
}

//
//
//

// PlainParser is a parser for config files in an extremely simple format. Each
// line is tokenized as a single key/value pair. The first space-delimited token
// in the line is interpreted as the flag name, and the rest of the line is
// interpreted as the flag value.
//
// Any leading hyphens on the flag name are ignored. Lines with a flag name but
// no value are interpreted as booleans, and the value is set to true.
//
// Flag values are trimmed of leading and trailing whitespace, but are otherwise
// unmodified. In particular, values are not quote-unescaped, and control
// characters like \n are not evaluated and instead passed through as literals.
//
// Comments are supported via "#". End-of-line comments require a space between
// the end of the line and the "#" character.
//
// An example config file follows.
//
//	# this is a full-line comment
//	timeout 250ms     # this is an end-of-line comment
//	foo     abc def   # set foo to `abc def`
//	foo     12345678  # repeated flags result in repeated calls to Set
//	bar     "abc def" # set bar to `"abc def"`, including quotes
//	baz     x\ny      # set baz to `x\ny`, passing \n literally
//	verbose           # equivalent to `verbose true`
func PlainParser(r io.Reader, set func(name, value string) error) error {
	s := bufio.NewScanner(r)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" {
			continue // skip empties
		}

		if line[0] == '#' {
			continue // skip comments
		}

		var (
			name  string
			value string
			index = strings.IndexRune(line, ' ')
		)
		if index < 0 {
			name, value = line, "true" // boolean option
		} else {
			name, value = line[:index], strings.TrimSpace(line[index:])
		}

		if i := strings.Index(value, " #"); i >= 0 {
			value = strings.TrimSpace(value[:i])
		}

		if err := set(name, value); err != nil {
			return err
		}
	}
	return s.Err()
}

//
//
//

var envVarSeparators = strings.NewReplacer(
	"-", "_",
	".", "_",
	"/", "_",
)

func getEnvVarKeys(f Flag, pc ParseContext) []string {
	var keys []string
	if longName, ok := f.GetLongName(); ok {
		keys = append(keys, getEnvVarKey(longName, pc))
	}
	if shortName, ok := f.GetShortName(); ok && pc.envVarShortNames {
		keys = append(keys, getEnvVarKey(string(shortName), pc))
	}
	return keys
}

func getEnvVarKey(flagName string, pc ParseContext) string {
	var key string
	key = flagName
	key = strings.TrimLeft(key, "-")
	key = envVarSeparators.Replace(key)
	if pc.envVarCaseSensitive {
		key = maybePrefix(key, pc.envVarPrefix)
	} else {
		key = maybePrefix(key, strings.ToUpper(pc.envVarPrefix))
		key = strings.ToUpper(key)
	}
	return key
}

func maybePrefix(key string, prefix string) string {
	if prefix != "" {
		key = prefix + "_" + key
	}
	return key
}

func splitEscape(s string, separator string) []string {
	escape := `\`
	tokens := strings.Split(s, separator)
	for i := len(tokens) - 2; i >= 0; i-- {
		if strings.HasSuffix(tokens[i], escape) {
			tokens[i] = tokens[i][:len(tokens[i])-len(escape)] + separator + tokens[i+1]
			tokens = append(tokens[:i+1], tokens[i+2:]...)
		}
	}
	return tokens
}

//
//
//

type flagSetSlice []Flag

func (fss *flagSetSlice) add(f Flag) {
	for _, ff := range *fss {
		if f == ff {
			return
		}
	}
	*fss = append(*fss, f)
}

func (fss *flagSetSlice) has(f Flag) bool {
	for _, ff := range *fss {
		if f == ff {
			return true
		}
	}
	return false
}
