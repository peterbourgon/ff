package ffval

import (
	"flag"
	"fmt"
	"reflect"
)

// Value is a generic [flag.Value] that can be set from a string.
//
// Most consumers should probably not need to use a value directly, and can
// instead use one of the specific value types defined by this package, like
// [String] or [Duration].
type Value[T any] struct {
	// ParseFunc parses a string to the type T. If no ParseFunc is provided, and
	// T is a supported [ValueType], then a default ParseFunc will be assigned
	// lazily. If no ParseFunc is provided, and T is not a supported
	// [ValueType], then most method calls will panic.
	ParseFunc func(string) (T, error)

	// Pointer is the actual instance of the type T which is managed and updated
	// by the value. If no Pointer is provided, a new T is allocated lazily. For
	// this reason, callers should generally access the pointer via GetPointer,
	// rather than reading the field directly.
	Pointer *T

	// Default value, which is the zero value of the type T by default.
	Default T

	initialized bool
	isSet       bool
}

var _ flag.Value = (*Value[any])(nil)

// NewValue returns a [Value] of underlying [ValueType] T, which updates the
// given pointer ptr when set, and which has a default value of the zero value
// of the type T.
func NewValue[T ValueType](ptr *T) *Value[T] {
	var zero T
	return NewValueDefault(ptr, zero)
}

// NewValueDefault returns a value of underlying [ValueType] T, which updates
// the given pointer ptr when set, and which has the given default value def.
func NewValueDefault[T ValueType](ptr *T, def T) *Value[T] {
	v := &Value[T]{
		Pointer: ptr,
		Default: def,
	}
	v.initialize()
	return v
}

// NewValueParser returns a value for any type T that can be parsed from a
// string.
//
// This constructor is intended as a convenience function for tests; consumers
// who want to provide a parser are probably better served by constructing a
// value directly, so that they can also provide other fields in a single
// motion.
func NewValueParser[T any](parseFunc func(string) (T, error)) *Value[T] {
	v := &Value[T]{
		ParseFunc: parseFunc,
	}
	v.initialize()
	return v
}

func (v *Value[T]) initialize() {
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
		v.Pointer = new(T)
	}

	*v.Pointer = v.Default

	v.initialized = true
}

// Set the value by parsing the given string.
func (v *Value[T]) Set(s string) error {
	v.initialize()

	val, err := v.ParseFunc(s)
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	*v.Pointer = val
	v.isSet = true
	return nil
}

// Get the current value.
func (v *Value[T]) Get() T {
	v.initialize()
	return *v.Pointer
}

// GetPointer returns a pointer to the underlying instance of type T which is
// managed by this value.
func (v *Value[T]) GetPointer() *T {
	v.initialize()
	return v.Pointer
}

// Reset the value to its default state.
func (v *Value[T]) Reset() error {
	v.initialize()
	*v.Pointer = v.Default
	v.isSet = false
	return nil
}

// String returns a string representation of the value returned by Get.
func (v *Value[T]) String() string {
	return fmt.Sprint(v.Get())
}

// IsSet returns true if the value has been explicitly set.
func (v Value[T]) IsSet() bool {
	return v.isSet
}

// IsBoolFlag returns true if the underlying type T is bool.
func (v Value[T]) IsBoolFlag() bool {
	switch x := any(v.Default); x.(type) {
	case bool:
		return true
	default:
		return false
	}
}

//
//
//

type reflectValue struct {
	set func(string) error
	get func() string

	isBoolFlag bool
	typeName   string
}

var _ flag.Value = (*reflectValue)(nil)

// NewValueReflect TODO
func NewValueReflect(typ reflect.Type, dst reflect.Value, def string) (flag.Value, error) {
	if !dst.CanSet() {
		return nil, fmt.Errorf("unassignable destination %s", dst.Type().Name())
	}

	parseFunc, ok := defaultParseFuncs[typ]
	if !ok {
		return nil, fmt.Errorf("unsupported type %s", typ)
	}

	set := func(s string) error {
		parseFuncVal := reflect.ValueOf(parseFunc)
		parseFuncIn := []reflect.Value{reflect.ValueOf(s)}
		parseFuncOut := parseFuncVal.Call(parseFuncIn)
		if len(parseFuncOut) != 2 {
			panic(fmt.Errorf("calling parseFunc: expected 2 values, got %d", len(parseFuncOut)))
		}
		if err, ok := parseFuncOut[1].Interface().(error); ok && err != nil {
			return err
		}
		dst.Set(parseFuncOut[0])
		return nil
	}

	get := func() string {
		sprintFuncVal := reflect.ValueOf(fmt.Sprint)
		sprintFuncIn := []reflect.Value{dst}
		sprintFuncOut := sprintFuncVal.Call(sprintFuncIn)
		if len(sprintFuncOut) != 1 {
			panic(fmt.Errorf("calling fmt.Sprint: expected 2 values, got %d", len(sprintFuncOut)))
		}
		var s string
		reflect.ValueOf(&s).Elem().Set(sprintFuncOut[0])
		return s
	}

	if def != "" {
		if err := set(def); err != nil {
			return nil, err
		}
	}

	isBoolFlag := typ.ConvertibleTo(reflect.TypeOf(*new(bool)))
	typeName := typ.Name()

	return &reflectValue{
		set:        set,
		get:        get,
		isBoolFlag: isBoolFlag,
		typeName:   typeName,
	}, nil
}

func (v *reflectValue) Set(s string) error  { return v.set(s) }
func (v *reflectValue) String() string      { return v.get() }
func (v *reflectValue) IsBoolFlag() bool    { return v.isBoolFlag }
func (v *reflectValue) GetTypeName() string { return v.typeName }
