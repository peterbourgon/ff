// Package ff provides a flags-first approach to runtime configuration.
//
// [Parse] populates a flag set with runtime configuration from environment
// variables and config files.
//
// [Command] can be used to build hierarchical command-line applications in a
// declarative style.
//
// [Flags] is the core interface of the package. Consumers may use the
// getopts(3)-inspired [CoreFlags] implementation, or provide their own.
package ff
