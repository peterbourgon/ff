// Package ff provides a flags-first approach to runtime configuration.
//
// The main function is [Parse], which mirrors [flag.FlagSet.Parse], populating
// a set of [Flags] from commandline arguments, environment variables, and/or a
// config file. [Option] values control parsing behavior.
//
// [NewFlags] provides a standard set of Flags, inspired by getopts(3). You can
// also parse a [flags.FlagSet] directly, or provide your own implementation of
// the Flags interface altogether.
//
// [Command] is also provided as a tool for building CLI applications, like
// docker or kubectl, in a simple and declarative style. It's intended to be
// easier to understand and maintain than more common alternatives.
package ff
