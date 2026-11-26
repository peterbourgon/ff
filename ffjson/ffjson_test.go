package ffjson_test

import (
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffjson"
	"github.com/peterbourgon/ff/v4/fftest"
)

func TestParser(t *testing.T) {
	t.Parallel()

	testcases := fftest.TestCases{
		{
			Name:       "empty input",
			ConfigFile: "testdata/empty.json",
			Want:       fftest.Vars{},
		},
		{
			Name:       "basic KV pairs",
			ConfigFile: "testdata/basic.json",
			Want:       fftest.Vars{S: "s", I: 10, B: true, D: 5 * time.Second},
		},
		{
			Name:       "value arrays",
			ConfigFile: "testdata/value_arrays.json",
			Want:       fftest.Vars{S: "bb", I: 12, B: true, D: 5 * time.Second, X: []string{"a", "B", "👍"}},
		},
		{
			Name:       "bad JSON file",
			ConfigFile: "testdata/bad.json",
			Want:       fftest.Vars{WantParseErrorIs: io.ErrUnexpectedEOF},
		},
		{
			Name:         "nested with '.'",
			ConfigFile:   "testdata/nested.json",
			Constructors: []fftest.Constructor{fftest.NewNestedConstructor(".")},
			Want:         fftest.Vars{S: "foo bar", I: 34, X: []string{"alpha", "beta", "delta"}},
		},
		{
			Name:         "nested with '-'",
			ConfigFile:   "testdata/nested.json",
			Constructors: []fftest.Constructor{fftest.NewNestedConstructor("-")},
			Options:      []ff.Option{ff.WithConfigFileParser(ffjson.Parser{Delimiter: "-"}.Parse)},
			Want:         fftest.Vars{S: "foo bar", I: 34, X: []string{"alpha", "beta", "delta"}},
		},
	}

	testcases.Run(t)
}

func TestParserRejectsTrailingData(t *testing.T) {
	t.Parallel()

	for _, input := range []string{
		`{"s":"first"} not-json`,
		`{"s":"first"} {"s":"second"}`,
	} {
		input := input
		t.Run(input, func(t *testing.T) {
			t.Parallel()
			setCalled := false
			err := ffjson.Parse(strings.NewReader(input), func(name, value string) error {
				setCalled = true
				return nil
			})
			if err == nil {
				t.Fatal("want error, have none")
			}
			if setCalled {
				t.Fatal("set called before the complete document was validated")
			}
		})
	}
}

func TestParserAcceptsTrailingWhitespace(t *testing.T) {
	t.Parallel()

	var name string
	err := ffjson.Parse(strings.NewReader("{\"s\":\"value\"} \n\t"), func(n, value string) error {
		name = n
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if name != "s" {
		t.Fatalf("want setter called with s, have %q", name)
	}
}

func TestParserReportsTrailingReadError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("read failed")
	setCalled := false
	err := ffjson.Parse(&failingReader{data: []byte(`{"s":"value"}`), err: wantErr}, func(name, value string) error {
		setCalled = true
		return nil
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("want %v, have %v", wantErr, err)
	}
	if setCalled {
		t.Fatal("set called before the complete document was validated")
	}
}

type failingReader struct {
	data []byte
	err  error
}

func (r *failingReader) Read(p []byte) (int, error) {
	if len(r.data) == 0 {
		return 0, r.err
	}
	n := copy(p, r.data)
	r.data = r.data[n:]
	return n, nil
}
