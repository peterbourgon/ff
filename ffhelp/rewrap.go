package ffhelp

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"unicode"
)

var (
	// DefaultWidth is the default number of columns that e.g. [Rewrap] will
	// use, if they can't be determined from the current TTY.
	DefaultWidth = 120

	ttyColumns = DefaultWidth
	getColumns = sync.Once{}
)

// Columns returns the number of columns in the current TTY, or [DefaultWidth]
// if the current TTY can't be determined.
func Columns() int {
	getColumns.Do(func() {
		if cols, err := sttySizeCols(); err == nil {
			ttyColumns = cols
		}
	})
	return ttyColumns
}

// Rewrap calls [RewrapAt] with [Columns] as the max width.
func Rewrap(s string) string {
	return RewrapAt(s, Columns())
}

// RewrapAt rewraps s at max columns.
func RewrapAt(s string, max int) string {
	var paragraphs []string
	for _, p := range strings.Split(s, "\n\n") {
		p := strings.TrimSpace(p)
		lines := strings.Split(p, "\n")
		for i := range lines {
			lines[i] = strings.TrimSpace(lines[i])
		}
		p = strings.Join(lines, "\n")
		if p == "" {
			continue
		}
		paragraphs = append(paragraphs, wrapString(p, max))
	}
	return strings.Join(paragraphs, "\n\n")
}

const nbsp = 0xA0

// Adapted from github.com/mitchellh/go-wordwrap.
func wrapString(s string, max int) string {
	var (
		output        strings.Builder
		currentLine   strings.Builder
		pendingSpaces strings.Builder
		pendingWord   strings.Builder
	)

	maybeWriteWord := func() {
		if pendingWord.Len() <= 0 {
			return
		}
		if tooLong := currentLine.Len()+pendingSpaces.Len()+pendingWord.Len() > max; tooLong {
			currentLine.WriteRune('\n')
			output.WriteString(currentLine.String())
			currentLine.Reset()
			pendingSpaces.Reset()
		}
		currentLine.WriteString(pendingSpaces.String())
		pendingSpaces.Reset()
		currentLine.WriteString(pendingWord.String())
		pendingWord.Reset()
	}

	for _, r := range s {
		if r == '\n' {
			r = ' '
		}
		switch {
		case unicode.IsSpace(r) && r != nbsp:
			maybeWriteWord()
			pendingSpaces.WriteRune(r)
		default:
			pendingWord.WriteRune(r)
		}
	}

	maybeWriteWord()

	if currentLine.Len() > 0 {
		output.WriteString(currentLine.String())
	}

	return output.String()
}

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

	if cols < 40 {
		cols = 40
	}

	if cols > 120 {
		cols = int(float64(cols) * 0.66)
	}

	return cols, nil
}
