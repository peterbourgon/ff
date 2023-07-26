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
	readEnvVars  bool
	envVarPrefix string
	envVarSplit  string

	configFileFilename     string
	configFileFlagName     string
	configFileParseFunc    ConfigFileParseFunc
	configFileOpenFunc     func(string) (iofs.File, error)
	allowMissingConfigFile bool
	ignoreUndefined        bool
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
		pc.configFileFilename = filename
	}
}

// WithConfigFileFlag tells [Parse] to treat the flag with the given name as a
// config file. The flag name must be defined in the flag set that parse
// consumes. Requires [WithConfigFileParser], and is overridden by
// [WithConfigFile]. To specify a default config file, provide it as the default
// value of the corresponding flag.
func WithConfigFileFlag(flagname string) Option {
	return func(pc *ParseContext) {
		pc.configFileFlagName = flagname
	}
}

// WithConfigFileParser tells [Parse] how to interpret a provided config file.
// This option is required to parse config files. If this option isn't provided,
// config files are ignored.
func WithConfigFileParser(pf ConfigFileParseFunc) Option {
	return func(pc *ParseContext) {
		pc.configFileParseFunc = pf
	}
}

// WithConfigAllowMissingFile tells [Parse] to ignore config files that are
// specified but don't exist.
//
// By default, missing config files result in a parse error.
func WithConfigAllowMissingFile() Option {
	return func(pc *ParseContext) {
		pc.allowMissingConfigFile = true
	}
}

// WithConfigIgnoreUndefinedFlags tells [Parse] to ignore undefined flags that
// it encounters in config files. This option only applies to flags in config
// files.
//
// By default, undefined flags in config files result in a parse error.
func WithConfigIgnoreUndefinedFlags() Option {
	return func(pc *ParseContext) {
		pc.ignoreUndefined = true
	}
}

// WithEnvVars tells [Parse] to set flags from environment variables. Flag names
// are matched to environment variables by capitalizing the flag name, and
// replacing separator characters like periods or hyphens with underscores.
//
// By default, flags are not parsed from environment variables at all.
func WithEnvVars() Option {
	return func(pc *ParseContext) {
		pc.readEnvVars = true
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
		pc.readEnvVars = true
		pc.envVarPrefix = prefix
	}
}

// WithEnvVarSplit tells [Parse] to split environment variable values on the
// given delimiter, and to set the flag multiple times, with each token as a
// distinct value. Values produced in this way are not trimmed of whitespace.
//
// For example, `FOO=a,b,c` might cause a flag named `foo` to receive a single
// call to Set with the value `a,b,c`. If the split delimiter were set to `,`
// then that flag would receive three seperate calls to Set with the strings
// `a`, `b`, and `c`.
//
// By default, no splitting of environment variable values occurs.
func WithEnvVarSplit(delimiter string) Option {
	return func(pc *ParseContext) {
		pc.readEnvVars = true
		pc.envVarSplit = delimiter
	}
}

// WithFilesystem tells [Parse] to use the provided filesystem when accessing
// files on disk, for example when reading a config file. By default, the host
// filesystem is used, via [os.Open].
func WithFilesystem(fs embed.FS) Option {
	return func(pc *ParseContext) {
		pc.configFileOpenFunc = fs.Open
	}
}
