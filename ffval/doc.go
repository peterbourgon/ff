// Package ffval provides common flag value types and helpers.
//
// The types defined by this package implement [flag.Value], and are intended to
// be used as values in an [ff.CoreFlagConfig].
//
// [Value] represents a single instance of any type T that can be parsed from a
// string. The package defines a set of values for common underlying types, like
// [Bool], [String], [Duration], etc.
//
// [List] and [UniqueList] represent a sequence of values of type T, where each
// call to Set (potentially) adds the value to the end of the list. The package
// defines a small set of lists for common underlying types, like [IntList],
// [StringSet], etc.
package ffval
