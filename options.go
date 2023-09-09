package ff

import (
	"embed"
	"io"
	iofs "io/fs"
)

// Option controls some aspect of parsing behavior.
type Option func(*ParseContext)

// ParseContext receives and maintains parse options.
type ParseContext struct {
	envVarEnabled bool
	envVarPrefix  string
	envVarSplit   string

	configFileName             string
	configFlagName             string
	configParseFunc            ConfigFileParseFunc
	configOpenFunc             func(string) (iofs.File, error)
	configAllowMissingFile     bool
	configIgnoreUndefinedFlags bool
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

// WithEnvVars tells [Parse] to set flags from environment variables. Flags are
// matched to environment variables by capitalizing the flag name, and replacing
// separator characters like periods or hyphens with underscores.
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
// token. Values produced in this way are not trimmed of whitespace.
//
// For example, the env var `FOO=a,b,c` would by default set a flag named `foo`
// one time, with the value `a,b,c`. Providing WithEnvVarSplit with a comma
// delimiter would set `foo` multiple times, with the values `a`, `b`, and `c`.
//
// By default, no splitting of environment variable values occurs.
func WithEnvVarSplit(delimiter string) Option {
	return func(pc *ParseContext) {
		pc.envVarEnabled = true
		pc.envVarSplit = delimiter
	}
}

// WithFilesystem tells [Parse] to use the provided filesystem when accessing
// files on disk, typically when reading a config file.
//
// By default, the host filesystem is used, via [os.Open].
func WithFilesystem(fs embed.FS) Option {
	return func(pc *ParseContext) {
		pc.configOpenFunc = fs.Open
	}
}
