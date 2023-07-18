package internal_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/peterbourgon/ff/v3/internal"
)

func TestTraverseMap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Name  string
		M     map[string]any
		Delim string
		Want  map[string]struct{}
	}{
		{
			Name: "single values",
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
				"s=foo":   {},
				"i=123":   {},
				"i64=123": {},
				"u=123":   {},
				"f=1.23":  {},
				"jn=123":  {},
				"b=true":  {},
				"nil=":    {},
			},
		},
		{
			Name: "slices",
			M: map[string]any{
				"is": []any{1, 2, 3},
				"ss": []any{"a", "b", "c"},
				"bs": []any{true, false},
				"as": []any{"a", 1, true},
			},
			Want: map[string]struct{}{
				"is=1":     {},
				"is=2":     {},
				"is=3":     {},
				"ss=a":     {},
				"ss=b":     {},
				"ss=c":     {},
				"bs=true":  {},
				"bs=false": {},
				"as=a":     {},
				"as=1":     {},
				"as=true":  {},
			},
		},
		{
			Name: "nested maps",
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
				"m.s=foo":    {},
				"m.m2.i=123": {},
			},
		},
		{
			Name: "nested maps with '-' delimiter",
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
				"m-s=foo":    {},
				"m-m2-i=123": {},
			},
		},
		{
			Name: "nested map[any]any",
			M: map[string]any{
				"m": map[any]any{
					"m2": map[string]any{
						"i": 999,
					},
				},
			},
			Delim: ".",
			Want: map[string]struct{}{
				"m.m2.i=999": {},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			observe := func(name, value string) error {
				key := name + "=" + value
				if _, ok := test.Want[key]; !ok {
					t.Errorf("set(%s, %s): unexpected call to set", name, value)
				}
				delete(test.Want, key)
				return nil
			}

			if err := internal.TraverseMap(test.M, test.Delim, observe); err != nil {
				t.Fatal(err)
			}

			for key := range test.Want {
				name, value, _ := strings.Cut(key, "=")
				t.Errorf("set(%s, %s): expected but did not occur", name, value)
			}
		})
	}
}
