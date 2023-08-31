// Package ff provides a flags-first approach to runtime configuration.
//
// [Parse] is the central function. It mirrors [flag.FlagSet.Parse] and
// populates a set of [Flags] from commandline arguments, environment variables,
// and/or a config file. [Option] values control parse behavior.
//
// [CoreFlags] is a standard, getopts(3)-inspired implementation of the [Flags]
// interface. Consumers can create a CoreFlags via [NewFlags], or adapt an
// existing [flag.FlagSet] to a CoreFlags via [NewStdFlags], or provide their
// own implementation altogether.
//
// [Command] is provided as a way to build hierarchical CLI tools, like docker
// or kubectl, in a simple and declarative style. It's intended to be easier to
// understand and maintain than more common alternatives.
package ff
