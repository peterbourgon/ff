package ffval_test

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/peterbourgon/ff/v4/ffval"
)

func TestLists_zero(t *testing.T) {
	t.Parallel()

	t.Run("List[int]", func(t *testing.T) {
		var val ffval.List[int]

		if want, have := (*[]int)(nil), val.Pointer; want != have {
			t.Errorf("Pointer: want %#+v, have %#+v", want, have) // nil Pointer is lazy-initialized
		}

		if want, have := []int{}, val.Get(); !reflect.DeepEqual(want, have) {
			t.Errorf("GetPointer: want %v, have %v", want, have)
		}

		if val.Pointer == nil {
			t.Fatalf("Pointer: still nil after GetPointer")
		}

		if want, have := "", val.String(); want != have {
			t.Errorf("String: want %q, have %q", want, have)
		}

		if want, have := []int{}, val.Get(); !reflect.DeepEqual(want, have) {
			t.Errorf("Get: want %#+v, have %#+v", want, have)
		}

		if err := val.Set("123"); err != nil {
			t.Fatalf("Set(123): %v", err)
		}

		if err := val.Set("123"); err != nil {
			t.Fatalf("Set(123): %v", err)
		}

		if want, have := "123, 123", val.String(); want != have {
			t.Errorf("String: want %q, have %q", want, have)
		}

		if want, have := []int{123, 123}, val.Get(); !reflect.DeepEqual(want, have) {
			t.Errorf("Get: want %v, have %v", want, have)
		}

		if want, have := []int{123, 123}, *val.Pointer; !reflect.DeepEqual(want, have) {
			t.Errorf("Pointer: want %v, have %v", want, have)
		}
	})

	t.Run("UniqueList[int]", func(t *testing.T) {
		var val ffval.UniqueList[int]

		if want, have := (*[]int)(nil), val.Pointer; want != have {
			t.Errorf("Pointer: want %#+v, have %#+v", want, have) // nil Pointer is lazy-initialized
		}

		if want, have := []int{}, val.Get(); !reflect.DeepEqual(want, have) {
			t.Errorf("GetPointer: want %v, have %v", want, have)
		}

		if val.Pointer == nil {
			t.Fatalf("Pointer: still nil after GetPointer")
		}

		if want, have := "", val.String(); want != have {
			t.Errorf("String: want %q, have %q", want, have)
		}

		if want, have := []int{}, val.Get(); !reflect.DeepEqual(want, have) {
			t.Errorf("Get: want %#+v, have %#+v", want, have)
		}

		if err := val.Set("123"); err != nil {
			t.Fatalf("Set(123): %v", err)
		}

		if err := val.Set("456"); err != nil {
			t.Fatalf("Set(456): %v", err)
		}

		if err := val.Set("123"); err != nil {
			t.Fatalf("Set(123): %v", err)
		}

		val.ErrDuplicate = fmt.Errorf("dupe")
		if err := val.Set("123"); !errors.Is(err, val.ErrDuplicate) {
			t.Fatalf("Set(123): want %v, have %v", val.ErrDuplicate, err)
		}

		if want, have := "123, 456", val.String(); want != have {
			t.Errorf("String: want %q, have %q", want, have)
		}

		if want, have := []int{123, 456}, val.Get(); !reflect.DeepEqual(want, have) {
			t.Errorf("Get: want %v, have %v", want, have)
		}

		if want, have := []int{123, 456}, *val.Pointer; !reflect.DeepEqual(want, have) {
			t.Errorf("Pointer: want %v, have %v", want, have)
		}
	})
}

func TestLists_reset(t *testing.T) {
	t.Parallel()

	var list ffval.List[string]
	var set ffval.UniqueList[string]

	for _, s := range []string{"a", "a", "b", "c"} {
		list.Set(s)
		set.Set(s)
	}

	if want, have := "a, a, b, c", list.String(); want != have {
		t.Errorf("StringList: want %q, have %q", want, have)
	}

	if want, have := "a, b, c", set.String(); want != have {
		t.Errorf("StringSet: want %q, have %q", want, have)
	}

	list.Reset()
	set.Reset()

	if want, have := "", list.String(); want != have {
		t.Errorf("StringList: want %q, have %q", want, have)
	}

	if want, have := "", set.String(); want != have {
		t.Errorf("StringSet: want %q, have %q", want, have)
	}
}

func TestEnum(t *testing.T) {
	t.Parallel()

	t.Run("0 valid", func(t *testing.T) {
		defer func() {
			if x := recover(); x == nil {
				t.Errorf("expected panic, got none")
			}
		}()
		ffval.NewEnum(new(string))
	})

	t.Run("1 valid", func(t *testing.T) {
		e := ffval.NewEnum(new(int), 32)
		if want, have := 32, e.GetDefault(); want != have {
			t.Errorf("GetDefault: want %d, have %d", want, have)
		}
		if want, have := 32, e.Get(); want != have {
			t.Errorf("Get: want %d, have %d", want, have)
		}
		if err := e.Set("32"); err != nil {
			t.Errorf("Set(32): %v", err)
		}
		if err := e.Set("64"); err == nil {
			t.Errorf("Set(64): want error, have none")
		}
		if want, have := 32, e.Get(); want != have {
			t.Errorf("Get: want %d, have %d", want, have)
		}
	})

	t.Run("3 valid", func(t *testing.T) {
		e := ffval.NewEnum(new(int), 32, 64, 0)
		if want, have := 0, e.GetDefault(); want != have {
			t.Errorf("GetDefault: want %d, have %d", want, have)
		}
		if want, have := 0, e.Get(); want != have {
			t.Errorf("Get: want %d, have %d", want, have)
		}
		if err := e.Set("64"); err != nil {
			t.Errorf("Set(64): %v", err)
		}
		if err := e.Set("99"); !errors.Is(err, ffval.ErrInvalidValue) {
			t.Errorf("Set(99): want %v, have %v", ffval.ErrInvalidValue, err)
		}
		if want, have := 64, e.Get(); want != have {
			t.Errorf("Get: want %d, have %d", want, have)
		}
		if err := e.Reset(); err != nil {
			t.Fatalf("Reset: %v", err)
		}
		if want, have := 0, e.Get(); want != have {
			t.Errorf("Get: want %d, have %d", want, have)
		}
	})

	t.Run("direct", func(t *testing.T) {
		var x string
		e := &ffval.Enum[string]{
			// ParseFunc should be lazily assigned
			Valid:   []string{"foo", "bar", "baz"},
			Pointer: &x,
			Default: "bar",
		}
		if want, have := "bar", e.GetDefault(); want != have {
			t.Errorf("GetDefault: want %q, have %q", want, have)
		}
		if want, have := "bar", e.Get(); want != have {
			t.Errorf("Get: want %q, have %q", want, have)
		}
		if err := e.Set("foo"); err != nil {
			t.Errorf("Set(foo): %v", err)
		}
		if want, have := "bar", e.GetDefault(); want != have {
			t.Errorf("GetDefault: want %q, have %q", want, have)
		}
		if want, have := "foo", e.Get(); want != have {
			t.Errorf("Get: want %q, have %q", want, have)
		}
	})

	t.Run("custom type", func(t *testing.T) {
		type myint int
		var x myint
		e := &ffval.Enum[myint]{
			ParseFunc: func(s string) (myint, error) { i, err := strconv.Atoi(s); return myint(i), err },
			Valid:     []myint{1, 2, 3},
			Pointer:   &x,
		}
		if want, have := myint(0), x; want != have { // zero value for now
			t.Errorf("x: want %v, have %v", want, have)
		}
		if want, have := myint(1), e.GetDefault(); want != have {
			t.Errorf("GetDefault: want %q, have %q", want, have)
		}
		if want, have := myint(1), e.Get(); want != have {
			t.Errorf("Get: want %v, have %v", want, have)
		}
		if want, have := myint(1), x; want != have { // after lazy init, x should be set to the default value
			t.Errorf("x: want %v, have %v", want, have)
		}
		if err := e.Set(""); err == nil {
			t.Errorf("Set(''): want error, have none")
		}
		if err := e.Set("abc"); err == nil {
			t.Errorf("Set(abc): want error, have none")
		}
		if err := e.Set("0"); err == nil {
			t.Errorf("Set(0): want error, have none")
		}
		if err := e.Set("2"); err != nil {
			t.Errorf("Set(2): %v", err)
		}
		if want, have := myint(2), e.Get(); want != have {
			t.Errorf("Get: want %v, have %v", want, have)
		}
		if want, have := myint(2), *e.GetPointer(); want != have {
			t.Errorf("GetPointer: want %v, have %v", want, have)
		}
		if want, have := myint(2), x; want != have {
			t.Errorf("x: want %v, have %v", want, have)
		}
		if want, have := "2", e.String(); want != have {
			t.Errorf("String: want %q, have %q", want, have)
		}
		if err := e.Reset(); err != nil {
			t.Errorf("Reset: %v", err)
		}
		if want, have := myint(1), e.Get(); want != have {
			t.Errorf("Get: want %v, have %v", want, have)
		}
	})
}
