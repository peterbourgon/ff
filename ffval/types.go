package ffval

import (
	"reflect"
	"strconv"
	"time"
)

// ValueType is a generic type constraint for a specific set of primitive types
// that are natively supported by this package. Each of them has a default
// parser, which will be used if a parser is not explicitly provided by the
// user. This permits the zero value of corresponding generic types to be
// useful, which in turn allows this package to provide common and useful types
// like [Bool], [Duration], [StringSet], etc.
type ValueType interface {
	bool | int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | float32 | float64 | string | complex64 | complex128 | time.Duration
}

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

// Duration is a flag value representing a [time.Duration].
// Values are parsed by [time.ParseDuration].
type Duration = Value[time.Duration]

// StringList is a flag value representing a sequence of strings.
type StringList = List[string]

// StringSet is a flag value representing a unique set of strings.
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
	reflect.TypeOf(*new(time.Duration)): time.ParseDuration,
}
