// Package ff provides a flags-first approach to runtime configuration.
//
// The central function is [Parse], which mirrors [flag.FlagSet.Parse],
// populating a flag set from commandline arguments, environment variables,
// and/or a config file. [Option] values control parsing behavior.
//
// [NewFlags] provides a standard flag set, inspired by getopts(3). You can also
// parse a [flag.FlagSet] directly, or provide your own implementation of the
// [Flags] interface altogether.
//
// [Command] is also provided as a tool for building CLI applications, like
// docker or kubectl, in a simple and declarative style. It's intended to be
// easier to understand and maintain than common alternatives.
package ff
