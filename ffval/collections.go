package ffval

import (
	"errors"
	"flag"
	"fmt"
	"reflect"
	"strings"
)

// DefaultStringFunc is used by [List] and [UniqueList] if no StringFunc is
// explicitly provided. Each value is rendered to a string via [fmt.Sprint], and
// the strings are joined via [strings.Join] with a separator of ", ".
func DefaultStringFunc[T any](vals []T) string {
	strs := make([]string, len(vals))
	for i := range vals {
		strs[i] = fmt.Sprint(vals[i])
	}
	return strings.Join(strs, ", ")
}

// List is a generic [flag.Value] that represents an ordered list of values.
// Every call to Set adds the successfully parsed value to the end of the list.
// To prevent duplicate values, see [UniqueList].
type List[T any] struct {
	// ParseFunc parses a string to the type T. If no ParseFunc is provided, and
	// T is a supported [ValueType], then a default ParseFunc will be assigned
	// lazily. If no ParseFunc is provided, and T is not a supported
	// [ValueType], then most method calls will panic.
	ParseFunc func(string) (T, error)

	// Pointer is the actual slice of type T which is managed and updated by the
	// list. If no Pointer is provided, a new slice is allocated lazily. For
	// this reason, callers should generally access the pointer via GetPointer,
	// rather than reading the field directly.
	Pointer *[]T

	// StringFunc is used by the String method to transform the underlying slice
	// of T to a string. If no StringFunc is provided, [DefaultStringFunc] is
	// used.
	StringFunc func([]T) string

	initialized bool
	isSet       bool
}

var _ flag.Value = (*List[any])(nil)

// NewList returns a list of underlying [ValueType] T, which updates the given
// pointer ptr when set.
func NewList[T ValueType](ptr *[]T) *List[T] {
	v := &List[T]{
		Pointer: ptr,
	}
	v.initialize()
	return v
}

// NewListParser returns a list of any type T that can be parsed from a string.
//
// This constructor is intended as a convenience function for tests; consumers
// who want to provide a parser are probably better served by constructing a
// list directly, so that they can also provide other fields in a single motion.
func NewListParser[T any](parseFunc func(string) (T, error)) *List[T] {
	v := &List[T]{
		ParseFunc: parseFunc,
	}
	v.initialize()
	return v
}

func (v *List[T]) initialize() {
	if v.initialized {
		return
	}

	if v.ParseFunc == nil {
		var zero T
		valueType := reflect.TypeOf(zero)
		parse, ok := defaultParseFuncs[valueType]
		if !ok {
			panic(fmt.Errorf("%s: unsupported value type", valueType.String()))
		}
		pf, ok := parse.(func(string) (T, error))
		if !ok {
			panic(fmt.Errorf("%s: invalid default parse func (%T)", valueType.String(), parse))
		}
		v.ParseFunc = pf
	}

	if v.Pointer == nil {
		v.Pointer = &([]T{})
	}

	if v.StringFunc == nil {
		v.StringFunc = DefaultStringFunc[T]
	}

	v.initialized = true
}

// Set parses the given string, and appends the successfully parsed value to the
// list. Duplicates are permitted.
func (v *List[T]) Set(s string) error {
	v.initialize()

	value, err := v.ParseFunc(s)
	if err != nil {
		return err
	}

	*v.Pointer = append(*v.Pointer, value)
	v.isSet = true
	return nil
}

// Get the current list of values.
func (v *List[T]) Get() []T {
	v.initialize()
	return *v.Pointer
}

// GetPointer returns a pointer to the underlying slice of T.
func (v *List[T]) GetPointer() *[]T {
	v.initialize()
	return v.Pointer
}

// Reset the list of values to its default (empty) state.
func (v *List[T]) Reset() error {
	v.initialize()
	*v.Pointer = (*v.Pointer)[:0]
	v.isSet = false
	return nil
}

// String returns a string representation of the list of values.
func (v *List[T]) String() string {
	v.initialize()
	return v.StringFunc(v.Get())
}

// IsSet returns true if the list has been explicitly set.
func (v *List[T]) IsSet() bool {
	return v.isSet
}

//
//
//

// UniqueList is a [List] that doesn't allow duplicate values.
type UniqueList[T comparable] struct {
	// ParseFunc parses a string to the type T. If no ParseFunc is provided, and
	// T is a supported [ValueType], then a default ParseFunc will be assigned
	// lazily. If no ParseFunc is provided, and T is not a supported
	// [ValueType], then most method calls will panic.
	ParseFunc func(string) (T, error)

	// Pointer is the actual slice of type T which is managed and updated by the
	// list. If no Pointer is provided, a new slice is allocated lazily.
	Pointer *[]T

	// StringFunc is used by the String method to transform the underlying slice
	// of T to a string. If no StringFunc is provided, [DefaultStringFunc] is
	// used.
	StringFunc func([]T) string

	// ErrDuplicate is returned by Set when it detects a duplicate value. By
	// default, ErrDuplicate is nil, so duplicate values are silently dropped.
	ErrDuplicate error

	initialized bool
	isSet       bool
}

var _ flag.Value = (*UniqueList[any])(nil)

// NewUniqueList returns a unique list of underlying [ValueType] T, which
// updates the given pointer ptr when set.
func NewUniqueList[T ValueType](ptr *[]T) *UniqueList[T] {
	v := &UniqueList[T]{
		Pointer: ptr,
	}
	v.initialize()
	return v
}

// NewUniqueListParser returns a unique list of any comparable type T that can
// be parsed from a string.
//
// This constructor is intended as a convenience function for tests; consumers
// who want to provide a parser are probably better served by constructing a
// unique list directly, so that they can also provide other fields in a single
// motion.
func NewUniqueListParser[T comparable](parseFunc func(string) (T, error)) *UniqueList[T] {
	v := &UniqueList[T]{
		ParseFunc: parseFunc,
	}
	v.initialize()
	return v
}

func (v *UniqueList[T]) initialize() {
	if v.initialized {
		return
	}

	if v.ParseFunc == nil {
		var zero T
		valueType := reflect.TypeOf(zero)
		parse, ok := defaultParseFuncs[valueType]
		if !ok {
			panic(fmt.Errorf("%s: unsupported value type", valueType.String()))
		}
		pf, ok := parse.(func(string) (T, error))
		if !ok {
			panic(fmt.Errorf("%s: invalid default parse func (%T)", valueType.String(), parse))
		}
		v.ParseFunc = pf
	}

	if v.Pointer == nil {
		v.Pointer = &([]T{})
	}

	if v.StringFunc == nil {
		v.StringFunc = DefaultStringFunc[T]
	}

	v.initialized = true
}

// Set parses the given string, and appends the successfully parsed value to the
// list. If the value already exists in the list, Set returns the UniqueList's
// ErrDuplicate field, which is nil by default.
func (v *UniqueList[T]) Set(s string) error {
	v.initialize()

	value, err := v.ParseFunc(s)
	if err != nil {
		return err
	}

	for _, existing := range *(v.Pointer) {
		if value == existing {
			return v.ErrDuplicate
		}
	}

	*v.Pointer = append(*v.Pointer, value)
	v.isSet = true
	return nil
}

// Get the current list of values.
func (v *UniqueList[T]) Get() []T {
	v.initialize()
	return *v.Pointer
}

// GetPointer returns a pointer to the underlying slice of T.
func (v *UniqueList[T]) GetPointer() *[]T {
	v.initialize()
	return v.Pointer
}

// Reset the list of values to its default (empty) state.
func (v *UniqueList[T]) Reset() error {
	v.initialize()
	*v.Pointer = (*v.Pointer)[:0]
	v.isSet = false
	return nil
}

// String returns a string representation of the list of values.
func (v *UniqueList[T]) String() string {
	v.initialize()
	return v.StringFunc(v.Get())
}

// IsSet returns true if the list has been explicitly set.
func (v *UniqueList[T]) IsSet() bool {
	return v.isSet
}

//
//
//

// Enum is a generic [flag.Value] that represents one of a fixed set of possible
// values of any comparable type T. An enum must have at least one valid value,
// or it is itself invalid. For this reason, the zero value of an enum is not
// useful, and calling most methods on a zero-value enum will panic.
type Enum[T comparable] struct {
	// ParseFunc parses a string to the type T. If no ParseFunc is provided, and
	// T is a supported [ValueType], then a default ParseFunc will be assigned
	// lazily. If no ParseFunc is provided, and T is not a supported
	// [ValueType], then most method calls will panic.
	ParseFunc func(string) (T, error)

	// Valid is the set of acceptable values. An enum with no valid values is
	// itself invalid, and all methods will panic.
	Valid []T

	// Pointer is the actual instance of the type T which is managed and updated
	// by the value. If no Pointer is provided, a new T is allocated lazily. For
	// this reason, callers should generally access the pointer via GetPointer,
	// rather than reading the field directly.
	Pointer *T

	// Default value of the enum. If the default value isn't valid, it will be
	// set to the first valid value lazily. For this reason, callers should
	// generally access the default via GetDefault, rather than reading the
	// field directly.
	Default T

	initialized bool
	isSet       bool
}

// ErrInvalidValue is returned when a value is set with invalid input.
var ErrInvalidValue = errors.New("invalid value")

var _ flag.Value = (*Enum[any])(nil)

// NewEnum returns an enum of [ValueType] T, updating the given pointer ptr when
// set, and which will accept only the provided valid values. At least one valid
// value is required, or else the function will panic.
func NewEnum[T ValueType](ptr *T, valid ...T) *Enum[T] {
	v := &Enum[T]{
		Pointer: ptr,
		Valid:   valid,
	}
	v.initialize()
	return v
}

// NewEnumParser returns an enum of any comparable type T that can be parsed
// from a string, and which will accept only the provided valid values. At least
// one valid value is required, or else the function will panic.
//
// This constructor is intended as a convenience function for tests; consumers
// who want to provide a parser are probably better served by constructing an
// enum directly, so that they can also provide other fields in a single motion.
func NewEnumParser[T comparable](parseFunc func(string) (T, error), valid ...T) *Enum[T] {
	v := &Enum[T]{
		ParseFunc: parseFunc,
		Valid:     valid,
	}
	v.initialize()
	return v
}

func (v *Enum[T]) initialize() {
	if v.initialized {
		return
	}

	if v.ParseFunc == nil {
		var zero T
		valueType := reflect.TypeOf(zero)
		parse, ok := defaultParseFuncs[valueType]
		if !ok {
			panic(fmt.Errorf("%s: unsupported value type", valueType.String()))
		}
		pf, ok := parse.(func(string) (T, error))
		if !ok {
			panic(fmt.Errorf("%s: invalid default parse func (%T)", valueType.String(), parse))
		}
		v.ParseFunc = pf
	}

	if len(v.Valid) <= 0 {
		panic(fmt.Errorf("no valid values provided"))
	}

	if v.Pointer == nil {
		v.Pointer = new(T)
	}

	var defaultValid bool
	for _, valid := range v.Valid {
		if v.Default == valid {
			defaultValid = true
			break
		}
	}
	if !defaultValid {
		v.Default = v.Valid[0]
	}

	*v.Pointer = v.Default

	v.initialized = true
}

// Set parses the given string, and sets the enum to the successfully parsed
// value, if it is valid. Otherwise, set returns ErrInvalidValue.
func (v *Enum[T]) Set(s string) error {
	v.initialize()

	value, err := v.ParseFunc(s)
	if err != nil {
		return err
	}

	for _, valid := range v.Valid {
		if value == valid {
			*v.Pointer = value
			v.isSet = true
			return nil
		}
	}

	return ErrInvalidValue
}

// Get the current value.
func (v *Enum[T]) Get() T {
	v.initialize()
	return *v.Pointer
}

// GetPointer returns a pointer to the underlying value.
func (v *Enum[T]) GetPointer() *T {
	v.initialize()
	return v.Pointer
}

// GetDefault returns the default value.
func (v *Enum[T]) GetDefault() T {
	v.initialize()
	return v.Default
}

// Reset the enum to its initial state.
func (v *Enum[T]) Reset() error {
	v.initialize()
	*v.Pointer = v.Default
	v.isSet = false
	return nil
}

// String returns a string representation of the current value.
func (v *Enum[T]) String() string {
	v.initialize()
	return fmt.Sprint(v.Get())
}

// IsSet returns true if the enum has been explicitly set.
func (v *Enum[T]) IsSet() bool {
	return v.isSet
}
