package ffval

import (
	"fmt"
	"reflect"
	"strconv"
	"time"
	"unicode/utf8"
)

// ValueType is a generic type constraint for a specific set of primitive types
// that are natively supported by this package. Each of them has a default
// parser, which will be used if a parser is not explicitly provided by the
// user. This permits the zero value of corresponding generic types to be
// useful, which in turn allows this package to provide common and useful types
// like [Bool], [String], [StringSet], etc.
type ValueType interface {
	bool | int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | float32 | float64 | string | complex64 | complex128 | time.Duration | ffbyte | ffrune
}

//
//
//

// Bool is a flag value representing a bool.
// Values are parsed by [strconv.ParseBool].
type Bool = Value[bool]

// Int is a flag value representing an int.
// Values are parsed by [strconv.Atoi].
type Int = Value[int]

// Int8 is a flag value representing an int8.
// Values are parsed by [strconv.ParseInt].
type Int8 = Value[int8]

// Int16 is a flag value representing an int16.
// Values are parsed by [strconv.ParseInt].
type Int16 = Value[int16]

// Int32 is a flag value representing an int32.
// Values are parsed by [strconv.ParseInt].
type Int32 = Value[int32]

// Int64 is a flag value representing an int64.
// Values are parsed by [strconv.ParseInt].
type Int64 = Value[int64]

// Uint is a flag value representing a uint.
// Values are parsed by [strconv.ParseUint].
type Uint = Value[uint]

// Uint8 is a flag value representing a uint8.
// Values are parsed by [strconv.ParseUint].
type Uint8 = Value[uint8]

// Uint16 is a flag value representing a uint16.
// Values are parsed by [strconv.ParseUint].
type Uint16 = Value[uint16]

// Uint32 is a flag value representing a uint32.
// Values are parsed by [strconv.ParseUint].
type Uint32 = Value[uint32]

// Uint64 is a flag value representing a uint64.
// Values are parsed by [strconv.ParseUint].
type Uint64 = Value[uint64]

// Float32 is a flag value representing a float32.
// Values are parsed by [strconv.ParseFloat].
type Float32 = Value[float32]

// Float64 is a flag value representing a float64.
// Values are parsed by [strconv.ParseFloat].
type Float64 = Value[float64]

// String is a flag value representing a string.
type String = Value[string]

// Complex64 is a flag value representing a complex64.
// Values are parsed by [strconv.ParseComplex].
type Complex64 = Value[complex64]

// Complex128 is a flag value representing a complex128.
// Values are parsed by [strconv.ParseComplex].
type Complex128 = Value[complex128]

// Byte is a flag value representing a byte. Values are parsed with a
// custom function that expects a string containing a single byte.
type Byte = Value[ffbyte]

// Rune is a flag value representing a rune. Values are parsed with a
// custom function that expects a string containing a single valid rune.
type Rune = Value[ffrune]

// Duration is a flag value representing a [time.Duration].
// Values are parsed by [time.ParseDuration].
type Duration = Value[time.Duration]

//
//
//

// BoolList is a [List] of bools. Duplicates are permitted.
type BoolList = List[bool]

// BoolSet is a [UniqueList] of bools. Duplicates are silently dropped.
type BoolSet = UniqueList[bool]

// IntList is a [List] of ints. Duplicates are permitted.
type IntList = List[int]

// IntSet is a [UniqueList] of ints. Duplicates are silently dropped.
type IntSet = UniqueList[int]

// StringList a [List] of strings. Duplicates are permitted.
type StringList = List[string]

// StringSet is a [UniqueList] of strings. Duplicates are silently dropped.
type StringSet = UniqueList[string]

//
//
//

var defaultParseFuncs = map[reflect.Type]any{
	reflect.TypeOf(*new(bool)):          strconv.ParseBool,
	reflect.TypeOf(*new(int)):           strconv.Atoi,
	reflect.TypeOf(*new(int8)):          func(s string) (int8, error) { v, err := strconv.ParseInt(s, 0, 8); return int8(v), err },
	reflect.TypeOf(*new(int16)):         func(s string) (int16, error) { v, err := strconv.ParseInt(s, 0, 16); return int16(v), err },
	reflect.TypeOf(*new(int32)):         func(s string) (int32, error) { v, err := strconv.ParseInt(s, 0, 32); return int32(v), err },
	reflect.TypeOf(*new(int64)):         func(s string) (int64, error) { v, err := strconv.ParseInt(s, 0, 64); return int64(v), err },
	reflect.TypeOf(*new(uint)):          func(s string) (uint, error) { v, err := strconv.ParseUint(s, 0, 64); return uint(v), err },
	reflect.TypeOf(*new(uint8)):         func(s string) (uint8, error) { v, err := strconv.ParseUint(s, 0, 8); return uint8(v), err },
	reflect.TypeOf(*new(uint16)):        func(s string) (uint16, error) { v, err := strconv.ParseUint(s, 0, 16); return uint16(v), err },
	reflect.TypeOf(*new(uint32)):        func(s string) (uint32, error) { v, err := strconv.ParseUint(s, 0, 32); return uint32(v), err },
	reflect.TypeOf(*new(uint64)):        func(s string) (uint64, error) { v, err := strconv.ParseUint(s, 0, 64); return uint64(v), err },
	reflect.TypeOf(*new(float32)):       func(s string) (float32, error) { v, err := strconv.ParseFloat(s, 32); return float32(v), err },
	reflect.TypeOf(*new(float64)):       func(s string) (float64, error) { v, err := strconv.ParseFloat(s, 64); return float64(v), err },
	reflect.TypeOf(*new(string)):        func(s string) (string, error) { return s, nil },
	reflect.TypeOf(*new(complex64)):     func(s string) (complex64, error) { v, err := strconv.ParseComplex(s, 64); return complex64(v), err },
	reflect.TypeOf(*new(complex128)):    func(s string) (complex128, error) { v, err := strconv.ParseComplex(s, 128); return complex128(v), err },
	reflect.TypeOf(*new(ffbyte)):        parseByte,
	reflect.TypeOf(*new(ffrune)):        parseRune,
	reflect.TypeOf(*new(time.Duration)): time.ParseDuration,
}

// byte aliases uint8, but we want to distinguish them when parsing.
type ffbyte byte

func parseByte(s string) (ffbyte, error) {
	if b, err := strconv.ParseUint(s, 0, 8); err == nil {
		return ffbyte(b), nil
	}

	if b := []byte(s); len(b) == 1 {
		return ffbyte(b[0]), nil
	}

	return 0, fmt.Errorf("invalid string %q", s)
}

// rune aliases int32, but we want to distinguish them when parsing.
type ffrune rune

func parseRune(s string) (ffrune, error) {
	if n := utf8.RuneCountInString(s); n != 1 {
		return 0, fmt.Errorf("invalid string: want 1 rune, have %d", n)
	}

	r, _ := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError {
		return 0, fmt.Errorf("invalid string: invalid rune")
	}

	return ffrune(r), nil
}
