// Package ffval provides common flag value types and helpers.
//
// The types defined by this package implement [flag.Value], and are intended to
// be used as values in a core flag config.
//
// [Value] represents a single instance of any type T that can be parsed from a
// string. The package defines a set of values for common underlying types, like
// [Bool], [String], [Duration], etc.
//
// [List] and [UniqueList] represent a sequence of values of type T, where each
// call to set adds a value to the end of the list. [Enum] represents one of a
// specific set of values of type T.
package ffval
