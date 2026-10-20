package ffenv_test

import (
	"bufio"
	"errors"
	"io"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/peterbourgon/ff/v4/ffenv"
)

func TestParseReaderErrors(t *testing.T) {
	t.Parallel()

	readErr := errors.New("config read failed")
	for _, tc := range []struct {
		name     string
		reader   io.Reader
		wantErr  error
		wantSets int
	}{
		{
			name:    "before first line",
			reader:  iotest.ErrReader(readErr),
			wantErr: readErr,
		},
		{
			name:     "after valid line",
			reader:   io.MultiReader(strings.NewReader("FOO=bar\n"), iotest.ErrReader(readErr)),
			wantErr:  readErr,
			wantSets: 1,
		},
		{
			name:    "line exceeds scanner limit",
			reader:  strings.NewReader("FOO=" + strings.Repeat("x", bufio.MaxScanTokenSize) + "\n"),
			wantErr: bufio.ErrTooLong,
		},
		{
			name:     "clean end of file",
			reader:   strings.NewReader("FOO=bar"),
			wantSets: 1,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sets := 0
			err := ffenv.Parse(tc.reader, func(name, value string) error {
				sets++
				if name != "FOO" || value != "bar" {
					t.Errorf("unexpected setting %q=%q", name, value)
				}
				return nil
			})
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("error: want %v, have %v", tc.wantErr, err)
			}
			if sets != tc.wantSets {
				t.Errorf("settings applied: want %d, have %d", tc.wantSets, sets)
			}
		})
	}
}
