package ffval_test

import (
	"errors"
	"fmt"
	"reflect"
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

	var list ffval.StringList
	var set ffval.StringSet

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
