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

// Parse the flag set with the provided args. [Option] values can be used to
// influence parse behavior. For example, options exist to read flags from
// environment variables, config files, etc.
//
// The fs parameter must be of type [FlagSet] or [*flag.FlagSet]. Any other type
// will result in an error.
func Parse(fs any, args []string, options ...Option) error {
	switch reified := fs.(type) {
	case *flag.FlagSet:
		return parseFlagSet(NewStdSet(reified), args, options...)
	case FlagSet:
		return parseFlagSet(reified, args, options...)
	default:
		return fmt.Errorf("unsupported flag set %T", fs)
	}
}

func parseFlagSet(fs FlagSet, args []string, options ...Option) error {
	// The parse context manages options.
	var pc ParseContext
	for _, option := range options {
		option(&pc)
	}

	// Index valid flags by env var key, to support .env config files (below).
	env2flag := map[string]Flag{}
	{
		if err := fs.WalkFlags(func(f Flag) error {
			for _, name := range getNameStrings(f) {
				key := getEnvVarKey(name, pc.envVarPrefix)
				if existing, ok := env2flag[key]; ok {
					return fmt.Errorf("%s: %w (%s)", getNameString(f), ErrDuplicateFlag, getNameString(existing))
				}
				env2flag[key] = f
			}
			return nil
		}); err != nil {
			return err
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
			return fmt.Errorf("parse arguments: %w", err)
		}

		markProvided()
	}

	// Second priority: the environment, i.e. the session.
	{
		if pc.readEnvVars {
			if err := fs.WalkFlags(func(f Flag) error {
				// If the flag has already been set, we can't do anything.
				if provided.has(f) {
					return nil
				}

				// Look in the environment for each of the flag names.
				for _, name := range getNameStrings(f) {
					// Transform the flag name to an env var key.
					key := getEnvVarKey(name, pc.envVarPrefix)

					// Look up the value from the environment.
					val := os.Getenv(key)
					if val == "" {
						continue
					}

					// The value may need to be split.
					vals := []string{val}
					if pc.envVarSplit != "" {
						vals = strings.Split(val, pc.envVarSplit)
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
		if pc.configFileFilename != "" {
			configFile = pc.configFileFilename
		}

		// Next, check the flag name.
		if configFile == "" && pc.configFileFlagName != "" {
			if f, ok := fs.GetFlag(pc.configFileFlagName); ok {
				configFile = f.GetValue()
			}
		}

		// If they didn't provide an open func, set the default.
		if pc.configFileOpenFunc == nil {
			pc.configFileOpenFunc = func(s string) (iofs.File, error) {
				return os.Open(s)
			}
		}

		// Config files require both a filename and a parser.
		var (
			haveConfigFile  = configFile != ""
			haveParser      = pc.configFileParseFunc != nil
			parseConfigFile = haveConfigFile && haveParser
		)
		if parseConfigFile {
			configFile, err := pc.configFileOpenFunc(configFile)
			switch {
			case err == nil:
				defer configFile.Close()
				if err := pc.configFileParseFunc(configFile, func(name, value string) error {
					// The parser calls us with a name=value pair. We want to
					// allow the name to be either the actual flag name, or its
					// env var representation (to support .env files).
					var (
						setFlag, fromSet = fs.GetFlag(name)
						envFlag, fromEnv = env2flag[name]
						target           Flag
					)
					switch {
					case fromSet:
						target = setFlag
					case !fromSet && fromEnv:
						target = envFlag
					case !fromSet && !fromEnv && pc.ignoreUndefined:
						return nil
					case !fromSet && !fromEnv && !pc.ignoreUndefined:
						return fmt.Errorf("%s: %w", name, ErrUnknownFlag)
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

			case errors.Is(err, iofs.ErrNotExist) && pc.allowMissingConfigFile:
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
// trimmed of whitespace and interpreted as the value.
//
// Any leading hyphens on flag names are ignored. Lines with a flag name but no
// value are interpreted as booleans and the value is set to true. Values are
// trimmed of leading and trailing whitespace but otherwise unmodified. In
// particular, values are not quote-unescaped, and control characters line \n
// are treated as literals.
//
// Comments are supported via "#". End-of-line comments require a space between
// the end of the line and the "#" character.
//
// An example config file follows.
//
//	# this is a full-line comment
//	timeout 250ms     # this is an end-of-line comment
//	foo     abc def   # set foo to `abc def`
//	foo     12345678  # flags can be repeated
//	bar     "abc def" # set bar to `"abc def"` i.e. including quotes
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

func getEnvVarKey(flagName, envVarPrefix string) string {
	var key string
	key = flagName
	key = strings.TrimLeft(key, "-")
	key = strings.ToUpper(key)
	key = envVarSeparators.Replace(key)
	key = maybePrefix(key, envVarPrefix)
	return key
}

func maybePrefix(key string, prefix string) string {
	if prefix != "" {
		key = strings.ToUpper(prefix) + "_" + key
	}
	return key
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
