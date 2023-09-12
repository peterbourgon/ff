package ffval_test

import (
	"flag"
	"fmt"
	"strconv"
	"testing"

	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffval"
)

func TestValue_zero(t *testing.T) {
	t.Parallel()

	t.Run("Int", func(t *testing.T) {
		var val ffval.Int

		if want, have := (*int)(nil), val.Pointer; want != have {
			t.Errorf("Pointer: want %#+v, have %#+v", want, have) // nil Pointer is lazy-initialized
		}

		if val.GetPointer() == nil {
			t.Fatalf("GetPointer: nil")
		}

		if val.Pointer == nil {
			t.Fatalf("Pointer: still nil after GetPointer")
		}

		if want, have := "0", val.String(); want != have {
			t.Errorf("String: want %q, have %q", want, have)
		}

		if want, have := 0, val.Get(); want != have {
			t.Errorf("Get: want %v, have %v", want, have)
		}

		if err := val.Set("123"); err != nil {
			t.Fatalf("Set(123): %v", err)
		}

		if want, have := "123", val.String(); want != have {
			t.Errorf("String: want %q, have %q", want, have)
		}

		if want, have := 123, val.Get(); want != have {
			t.Errorf("Get: want %v, have %v", want, have)
		}

		if want, have := 123, *val.Pointer; want != have {
			t.Errorf("Pointer: want %v, have %v", want, have)
		}
	})
}

func TestValue_reset(t *testing.T) {
	t.Parallel()

	t.Run("String", func(t *testing.T) {
		val := ffval.Value[string]{
			ParseFunc: func(s string) (string, error) { return s, nil },
			Default:   "zombo.com",
		}

		if want, have := "zombo.com", val.Get(); want != have {
			t.Errorf("Get: want %q, have %q", want, have)
		}

		if err := val.Set("party.pizza"); err != nil {
			t.Fatalf("Set: %v", err)
		}

		if want, have := "party.pizza", val.Get(); want != have {
			t.Errorf("Get: want %q, have %q", want, have)
		}

		if err := val.Reset(); err != nil {
			t.Fatalf("Reset: %v", err)
		}

		if want, have := "zombo.com", val.Get(); want != have {
			t.Errorf("Get: want %q, have %q", want, have)
		}
	})
}

func TestValue_constructors(t *testing.T) {
	t.Parallel()

	t.Run("NewValueParser float64", func(t *testing.T) {
		val := ffval.NewValueParser(func(s string) (float64, error) {
			return strconv.ParseFloat(s, 64)
		})

		if want, have := 0.0, val.Get(); want != have {
			t.Errorf("Get: want %v, have %v", want, have)
		}

		if err := val.Set("1.23"); err != nil {
			t.Fatalf("Set(1.23): %v", err)
		}

		if want, have := 1.23, val.Get(); want != have {
			t.Errorf("Get: want %v, have %v", want, have)
		}
	})

	t.Run("NewValueParser uintptr", func(t *testing.T) {
		val := ffval.NewValueParser(func(s string) (uintptr, error) {
			u, err := strconv.ParseUint(s, 10, 64)
			return uintptr(u), err
		})

		if want, have := uintptr(0), val.Get(); want != have {
			t.Errorf("Get: want %v, have %v", want, have)
		}

		if err := val.Set("12345678"); err != nil {
			t.Fatalf("Set: %v", err)
		}

		if want, have := uintptr(12345678), val.Get(); want != have {
			t.Errorf("Get: want %v, have %v", want, have)
		}
	})
}

func TestValue_types(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		value flag.Value
		good  []string
		bad   []string
	}{
		{
			value: new(ffval.Bool),
			good:  []string{"1", "true", "TRUE", "True", "T", "t", "0", "false", "FALSE", "False", "F", "f"},
			bad:   []string{"", "yes", "2", "no"},
		},
		{
			value: new(ffval.Int),
			good:  []string{"0", "1", "-2", "123"},
			bad:   []string{"", "1e3", "999999999999999999999999", "0b01", "0o4", "0x9", "0xEf", "0XA0"},
		},
		{
			value: new(ffval.Int8),
			good:  []string{"0", "1", "-2", "0b01", "0o2", "0x03", "0X04", "0xa"},
			bad:   []string{"", "1e3", "xxx", "32768", "0xfa", "0XAF"},
		},
		{
			value: new(ffval.Int16),
			good:  []string{"0", "1", "-2", "0b01", "0o2", "0x03", "0xfa", "0XAF"},
			bad:   []string{"", "1e3", "xxx", "99999"},
		},
		{
			value: new(ffval.Int32),
			good:  []string{"0", "1", "-2", "0b01", "0o2", "0x03", "0xfa", "0XAF"},
			bad:   []string{"", "1e3", "xxx", "123456789012345"},
		},
		{
			value: new(ffval.Int64),
			good:  []string{"0", "1", "-2", "0b01", "0o2", "0x03", "0xfa", "0XAF"},
			bad:   []string{"", "1e3", "xxx", "999999999999999999999999"},
		},
		{
			value: new(ffval.Uint),
			good:  []string{"0", "1", "0b01", "0o2", "0x03", "0xfa", "0XAF"},
			bad:   []string{"", "1e2", "xxx", "-3"},
		},
		{
			value: new(ffval.Uint8),
			good:  []string{"0", "1", "255", "0b01", "0o2", "0x03", "0xfa", "0XAF"},
			bad:   []string{"", "1e2", "xxx", "-4", "256"},
		},
		{
			value: new(ffval.Uint16),
			good:  []string{"0", "1", "65535", "0b01", "0o2", "0x03", "0xfa", "0XAF"},
			bad:   []string{"", "1e2", "xxx", "-5", "65536"},
		},
		{
			value: new(ffval.Uint32),
			good:  []string{"0", "1", "4294967295", "0b01", "0o2", "0x03", "0xfa", "0XAF"},
			bad:   []string{"", "1e2", "xxx", "-6", "4294967296"},
		},
		{
			value: new(ffval.Uint64),
			good:  []string{"0", "1", "18446744073709551615", "0b01", "0o2", "0x03", "0xfa", "0XAF"},
			bad:   []string{"", "1e2", "xxx", "-7", "18446744073709551616"},
		},
		{
			value: new(ffval.Float32),
			good:  []string{"0", "-1", "-2.34", "5.6", "1e3"},
			bad:   []string{"", "xxx", "1e100", "1e500"},
		},
		{
			value: new(ffval.Float64),
			good:  []string{"0", "-1", "-2.34", "5.6", "1e3", "1e100"},
			bad:   []string{"", "xxx", "1e500"},
		},
		{
			value: new(ffval.String),
			good:  []string{"", "1", "hello", "🙂"},
			bad:   []string{},
		},
		{
			value: new(ffval.Complex64),
			good:  []string{"1", "(0)", "Inf", "+Inf", "-inf", "0.1i", "0x0p+012345i"},
			bad:   []string{"", " ", "i", "1e309i", "2e307"},
		},
		{
			value: new(ffval.Complex128),
			good:  []string{"1", "(0)", "Inf", "+Inf", "-inf", "0.1i", "0x0p+012345i", "2e307"},
			bad:   []string{"", " ", "i", "1e309i"},
		},
		{
			value: new(ffval.Duration),
			good:  []string{"12ns", "34ms", "5h6m", "127h"},
			bad:   []string{"", " ", "123", "3.21"},
		},
	} {
		t.Run(fmt.Sprintf("%T", test.value), func(t *testing.T) {
			fs := ff.NewFlagSet(t.Name())
			fs.Value('v', "value", test.value, "usage string")

			for _, s := range test.good {
				if err := test.value.Set(s); err != nil {
					t.Errorf("%T: %q: %v", test.value, s, err)
				}
			}

			for _, s := range test.bad {
				if err := test.value.Set(s); err == nil {
					t.Errorf("%T: %q: want error, have none", test.value, s)
				}
			}
		})
	}

	t.Run("bool", func(t *testing.T) {
		var b ffval.Bool
		if want, have := true, b.IsBoolFlag(); want != have {
			t.Errorf("%T: IsBoolFlag: want %v, have %v", b, want, have)
		}
	})
}

func TestValueReflect(t *testing.T) {
	t.Parallel()

	t.Run("string", func(t *testing.T) {
		var x string

		f, err := ffval.NewValueReflect(&x, "abc")
		if err != nil {
			t.Fatal(err)
		}

		if want, have := "abc", f.String(); want != have {
			t.Errorf("String: want %q, have %q", want, have)
		}

		if want, have := "abc", x; want != have {
			t.Errorf("x: want %q, have %q", want, have)
		}

		if err := f.Set("def"); err != nil {
			t.Errorf("Set(def): %v", err)
		}

		if want, have := "def", f.String(); want != have {
			t.Errorf("String: want %q, have %q", want, have)
		}

		if want, have := "def", x; want != have {
			t.Errorf("x: want %q, have %q", want, have)
		}
	})

	t.Run("float64", func(t *testing.T) {
		var x float64

		f, err := ffval.NewValueReflect(&x, "1.23")
		if err != nil {
			t.Fatal(err)
		}

		if want, have := "1.23", f.String(); want != have {
			t.Errorf("String: want %q, have %q", want, have)
		}

		if want, have := 1.23, x; want != have {
			t.Errorf("x: want %v, have %v", want, have)
		}

		if err := f.Set("4.56"); err != nil {
			t.Errorf("Set(def): %v", err)
		}

		if want, have := "4.56", f.String(); want != have {
			t.Errorf("String: want %q, have %q", want, have)
		}

		if want, have := 4.56, x; want != have {
			t.Errorf("x: want %v, have %v", want, have)
		}
	})

	t.Run("float64 invalid init", func(t *testing.T) {
		var x float64

		_, err := ffval.NewValueReflect(&x, "abc")
		if err == nil {
			t.Fatalf("expected error, got none")
		}
	})

	t.Run("float64 invalid set", func(t *testing.T) {
		var x float64

		f, err := ffval.NewValueReflect(&x, "")
		if err != nil {
			t.Fatal(err)
		}

		if want, have := "0", f.String(); want != have {
			t.Errorf("String: want %q, have %q", want, have)
		}

		if want, have := 0.0, x; want != have {
			t.Errorf("x: want %v, have %v", want, have)
		}

		if err := f.Set("x"); err == nil {
			t.Errorf("Set(x): want error, have none")
		}
	})
}
