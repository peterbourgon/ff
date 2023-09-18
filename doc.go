// Package ff provides a flags-first approach to runtime configuration.
//
// The central function is [Parse], which populates a flag set from commandline
// arguments, environment variables, and/or config files. Parse takes either an
// implementation of the [Flags] interface, or (for compatibility reasons) a
// concrete [flag.FlagSet]. [Option] values can be used to control parsing
// behavior.
//
// [FlagSet] is provided as a standard implementation of the [Flags] interface,
// inspired by getopts(3). Callers are free to provide their own implementation.
//
// [Command] is provided as a tool for building CLI applications, like docker or
// kubectl, in a simple and declarative style. It's intended to be easier to
// understand and maintain than common alternatives.
package ff
