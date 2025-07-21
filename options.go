package ff

import (
	"io"
	iofs "io/fs"
)

// Option controls some aspect of parsing behavior.
type Option func(*ParseContext)

// ParseContext receives and maintains parse options.
type ParseContext struct {
	envVarEnabled       bool
	envVarPrefix        string
	envVarSplit         string
	envVarCaseSensitive bool
	envVarShortNames    bool

	configFileName             string
	configFlagName             string
	configParseFunc            ConfigFileParseFunc
	configOpenFunc             func(string) (iofs.File, error)
	configAllowMissingFile     bool
	configIgnoreUndefinedFlags bool
	configKeyIgnoreFlagNames   bool
	configKeyIgnoreEnvVars     bool
}

// ConfigFileParseFunc is a function that consumes the provided reader as a config
// file, and calls the provided set function for every name=value pair it
// discovers.
type ConfigFileParseFunc func(r io.Reader, set func(name, value string) error) error

// WithConfigFile tells [Parse] to read the provided filename as a config file.
// Requires [WithConfigFileParser], and overrides [WithConfigFileFlag].
//
// Because config files should generally be user-specifiable, this option should
// rarely be used; prefer [WithConfigFileFlag].
func WithConfigFile(filename string) Option {
	return func(pc *ParseContext) {
		pc.configFileName = filename
	}
}

// WithConfigFileFlag tells [Parse] to treat the flag with the given name as a
// config file. The flag name must be defined in the flag set consumed by parse.
// Requires [WithConfigFileParser], and is overridden by [WithConfigFile].
//
// To specify a default config file, provide it as the default value of the
// corresponding flag.
func WithConfigFileFlag(flagname string) Option {
	return func(pc *ParseContext) {
		pc.configFlagName = flagname
	}
}

// WithConfigFileParser tells [Parse] how to interpret a config file. This
// option must be explicitly provided in order to parse config files.
//
// By default, no config file parser is defined, and config files are ignored.
func WithConfigFileParser(pf ConfigFileParseFunc) Option {
	return func(pc *ParseContext) {
		pc.configParseFunc = pf
	}
}

// WithFilesystem tells [Parse] to use the provided filesystem when accessing
// files on disk, typically when reading a config file. This can be useful when
// working with embedded filesystems.
//
// By default, the host filesystem is used, via [os.Open].
func WithFilesystem(fs iofs.FS) Option {
	return func(pc *ParseContext) {
		pc.configOpenFunc = fs.Open
	}
}

// WithConfigAllowMissingFile tells [Parse] to ignore config files that are
// specified but don't exist.
//
// By default, missing config files result in a parse error.
func WithConfigAllowMissingFile() Option {
	return func(pc *ParseContext) {
		pc.configAllowMissingFile = true
	}
}

// WithConfigIgnoreUndefinedFlags tells [Parse] to ignore flags in config files
// which are not defined in the parsed flag set. This option only applies to
// flags in config files.
//
// By default, undefined flags in config files result in a parse error.
func WithConfigIgnoreUndefinedFlags() Option {
	return func(pc *ParseContext) {
		pc.configIgnoreUndefinedFlags = true
	}
}

// WithConfigIgnoreFlagNames tells [Parse] to ignore the short and long names,
// when matching keys from config files to valid flags.
//
// By default, config file keys are matched to flags by short name, long name,
// and/or environment variable name(s).
//
// In practice, this option is really only useful when parsing .env files, to
// avoid ambiguity or unintentional matches.
func WithConfigIgnoreFlagNames() Option {
	return func(pc *ParseContext) {
		pc.configKeyIgnoreFlagNames = true
	}
}

// WithConfigIgnoreEnvVars tells [Parse] to ignore environment variables, when
// matching keys from config files to valid flags.
//
// By default, config file keys are matched to flags by short name, long name,
// and/or environment variable name(s).
//
// This option can be useful in situations where invalid config file keys need
// to be reliably detected and rejected, as it helps to prevent unintentional
// matches from the environment.
func WithConfigIgnoreEnvVars() Option {
	return func(pc *ParseContext) {
		pc.configKeyIgnoreEnvVars = true
	}
}

// WithEnvVars tells [Parse] to set flags from environment variables.
//
// Flags are matched to environment variables by capitalizing the flag's long
// name, and replacing separator characters like periods or hyphens with
// underscores. For example, the flag `-f, --foo-bar` would match the
// environment variable `FOO_BAR`.
//
// By default, flags are not parsed from environment variables at all.
func WithEnvVars() Option {
	return func(pc *ParseContext) {
		pc.envVarEnabled = true
	}
}

// WithEnvVarPrefix is like [WithEnvVars], but only considers environment
// variables beginning with the given prefix followed by an underscore. That
// prefix (and underscore) are removed before matching the env var key to a flag
// name. For example, the env var prefix `MYPROG` would mean that the env var
// `MYPROG_FOO` matches a flag named `foo`.
//
// By default, flags are not parsed from environment variables at all.
func WithEnvVarPrefix(prefix string) Option {
	return func(pc *ParseContext) {
		pc.envVarEnabled = true
		pc.envVarPrefix = prefix
	}
}

// WithEnvVarSplit tells [Parse] to split environment variable values on the
// given delimiter, and to set the flag multiple times, once for each delimited
// token. Values produced in this way are not trimmed of whitespace. By default,
// no splitting of environment variable values occurs.
//
// For example, the env var `FOO=a,b,c` would by default set a flag named `foo`
// one time, with the value `a,b,c`. Providing WithEnvVarSplit with a comma
// delimiter would set `foo` multiple times, with the values `a`, `b`, and `c`.
//
// If an env var value contains the delimiter prefixed by a single backslash,
// that occurrence will be treated as a literal string, and not as a split
// point. For example, `FOO=a,b\,c` with a delimiter of `,` would yield values
// `a` and `b,c`. Or, `FOO=axxxb\xxxc` with a delimiter of `xxx` would yield
// values `a` and `bxxxc`.
//
// For historical reasons, WithEnvVarSplit automatically enables environment
// variable parsing. This will change in a future release. Callers should always
// explicitly provide [WithEnvVars] or [WithEnvVarPrefix] to parse flags from
// the environment.
func WithEnvVarSplit(delimiter string) Option {
	return func(pc *ParseContext) {
		pc.envVarSplit = delimiter
	}
}

// WithEnvVarCaseSensitive tells [Parse] to match flags to environment variables
// using the exact case of the flag name, rather than the default behavior fo
// transforming the flag name to uppercase.
//
// For example, using WithEnvVarPrefix("MYPREFIX"), the flag `foo` would
// normally match the environment variable `MYPREFIX_FOO`. With this option, it
// would instead match the environment variable `MYPREFIX_foo`.
//
// WithEnvVarCaseSensitive does NOT automatically enable environment variable
// parsing. Callers must explicitly provide [WithEnvVars] or [WithEnvVarPrefix]
// to parse flags from the environment.
func WithEnvVarCaseSensitive() Option {
	return func(pc *ParseContext) {
		pc.envVarCaseSensitive = true
	}
}

// WithEnvVarShortNames tells [Parse] to match flags to environment variables
// using the short name of the flag in addition to the long name.
//
// For example, if a flag is defined as `-f, --foo`, then normally only the
// environment variable `FOO` would match. Using this option means that the
// environment variable `F` would also match.
//
// WithEnvVarShortNames does NOT automatically enable environment variable
// parsing. Callers must explicitly provide [WithEnvVars] or [WithEnvVarPrefix]
// to parse flags from the environment.
func WithEnvVarShortNames() Option {
	return func(pc *ParseContext) {
		pc.envVarShortNames = true
	}
}
