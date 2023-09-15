package ffhelp

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

// Rewrap calls [RewrapAt] with a max width of [Columns]. If columns is less
// than 40, the max width will be set to 40. If columns is greater than 180, the
// max width will be scaled down proportional to its value.
func Rewrap(s string) string {
	cols := Columns()
	switch {
	case cols < 40:
		cols = 40
	case cols > 180:
		cols = 180 + int(0.5*float64(cols-180))
	}
	return RewrapAt(s, cols)
}

// RewrapAt rewraps s at max columns. Each line in s is trimmed of any leading
// tabs. Single newlines are treated as spaces. Two or more newlines are treated
// as paragraph breaks.
func RewrapAt(s string, max int) string {
	var (
		pendingLine strings.Builder
		outputLines strings.Builder
	)

	writeField := func(field string) {
		switch {
		case pendingLine.Len()+1+len(field) > max: // need linebreak
			outputLines.WriteString(pendingLine.String())
			outputLines.WriteRune('\n')
			pendingLine.Reset()
			pendingLine.WriteString(field)
		case pendingLine.Len() > 0: // need a space first
			pendingLine.WriteRune(' ')
			pendingLine.WriteString(field)
		default: // just the field
			pendingLine.WriteString(field)
		}
	}

	for _, paragraph := range strings.Split(s, "\n\n") {
		paragraph := strings.TrimSpace(paragraph)
		if paragraph == "" {
			continue
		}

		for _, line := range strings.Split(paragraph, "\n") {
			line = strings.Trim(line, "\t") // allow strings to be defined in nested source
			if line == "" {
				continue
			}
			for _, field := range strings.Fields(line) {
				writeField(field)
			}
		}

		outputLines.WriteString(pendingLine.String())
		pendingLine.Reset()

		outputLines.WriteRune('\n')
		outputLines.WriteRune('\n')
	}

	return strings.TrimSuffix(outputLines.String(), "\n\n")
}

//
//
//

// DefaultColumns is the default value returned by [Columns].
var DefaultColumns = 120

// Columns in the current TTY, or [DefaultColumns] if the TTY can't be determined.
func Columns() int {
	getColumns.Do(func() {
		// Don't run the stty subprocess unless we have to.
		if cols, err := sttySizeCols(); err == nil {
			ttyColumns.Store(int64(cols))
		} else {
			ttyColumns.Store(int64(DefaultColumns))
		}
	})
	return int(ttyColumns.Load())
}

var (
	getColumns sync.Once
	ttyColumns atomic.Int64
)

func sttySizeCols() (int, error) {
	stty, err := exec.LookPath("stty")
	if err != nil {
		return 0, err
	}

	cmd := exec.Command(stty, "size")
	cmd.Stdin = os.Stdin
	out, err := cmd.CombinedOutput()
	if err != nil {
		return 0, err
	}

	fields := bytes.Fields(out)
	if len(fields) != 2 {
		return 0, fmt.Errorf("unexpected output (%s)", string(out))
	}

	cols, err := strconv.Atoi(string(fields[1]))
	if err != nil {
		return 0, fmt.Errorf("unexpected output (%s): %w", string(out), err)
	}

	return cols, nil
}
