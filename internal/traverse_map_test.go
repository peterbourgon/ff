package internal_test

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/peterbourgon/ff/v3/internal"
)

func TestTraverseMap(t *testing.T) {
	tests := []struct {
		M     map[string]any
		Delim string
		Want  map[string]struct{}
	}{
		{ // single values
			M: map[string]any{
				"s":   "foo",
				"i":   123,
				"i64": int64(123),
				"u":   uint64(123),
				"f":   1.23,
				"jn":  json.Number("123"),
				"b":   true,
				"nil": nil,
			},
			Delim: ".",
			Want: map[string]struct{}{
				"s:foo":   {},
				"i:123":   {},
				"i64:123": {},
				"u:123":   {},
				"f:1.23":  {},
				"jn:123":  {},
				"b:true":  {},
				"nil:":    {},
			},
		},
		{ // slices
			M: map[string]any{
				"is": []any{1, 2, 3},
				"ss": []any{"a", "b", "c"},
				"bs": []any{true, false},
				"as": []any{"a", 1, true},
			},
			Want: map[string]struct{}{
				"is:1":     {},
				"is:2":     {},
				"is:3":     {},
				"ss:a":     {},
				"ss:b":     {},
				"ss:c":     {},
				"bs:true":  {},
				"bs:false": {},
				"as:a":     {},
				"as:1":     {},
				"as:true":  {},
			},
		},
		{ // nested maps
			M: map[string]any{
				"m": map[string]any{
					"s": "foo",
					"m2": map[string]any{
						"i": 123,
					},
				},
			},
			Delim: ".",
			Want: map[string]struct{}{
				"m.s:foo":    {},
				"m.m2.i:123": {},
			},
		},
		{ // nested maps with "-" delimiter
			M: map[string]any{
				"m": map[string]any{
					"s": "foo",
					"m2": map[string]any{
						"i": 123,
					},
				},
			},
			Delim: "-",
			Want: map[string]struct{}{
				"m-s:foo":    {},
				"m-m2-i:123": {},
			},
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			// deleteSet deletes the key from test.WantSet if it exists.
			deleteSet := func(name, value string) error {
				key := name + ":" + value
				delete(test.Want, key)
				return nil
			}

			if err := internal.TraverseMap(test.M, test.Delim, deleteSet); err != nil {
				t.Fatal(err)
			}
			if len(test.Want) > 0 {
				t.Fatalf("Failed to match keys: %v", test.Want)
			}
		})
	}
}
