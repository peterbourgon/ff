package ff

import "time"

// Value represents a single configuration parameter.
// It must be implemented by each supported, concrete type.
type Value interface {
	Set(string) error
	Get() interface{}
	String() string
}

// StringValue wraps string and implements the Value interface.
type StringValue string

func newStringValue(def string, p *string) *StringValue {
	*p = def
	return (*StringValue)(p)
}

// Set implements Value.
func (v *StringValue) Set(val string) error {
	*v = StringValue(val)
	return nil
}

// Get implements Value.
func (v *StringValue) Get() interface{} { return string(*v) }

// String implements Value.
func (v *StringValue) String() string { return string(*v) }

// DurationValue wraps time.Duration and implements the Value interface.
type DurationValue time.Duration

func newDurationValue(def time.Duration, p *time.Duration) *DurationValue {
	*p = def
	return (*DurationValue)(p)
}

// Set implements Value.
func (v *DurationValue) Set(s string) error {
	parsed, err := time.ParseDuration(s)
	*v = DurationValue(parsed)
	return err
}

// Get implements Value.
func (v *DurationValue) Get() interface{} { return time.Duration(*v) }

// String implements Value.
func (v *DurationValue) String() string { return (*time.Duration)(v).String() }
