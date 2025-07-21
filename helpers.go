package ff

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

func newFlagError(f Flag, err error) error {
	return fmt.Errorf("%s: %w", getNameString(f), err)
}

func isValidShortName(short rune) bool {
	var (
		isZero  = short == 0
		isError = short == utf8.RuneError
		isValid = !isZero && !isError
	)
	return isValid
}

var badStuff = strings.Join([]string{
	string(rune(0x00)),   // zero rune
	string([]byte{0x00}), // zero/NUL byte
	` `,                  // space
	"\t\n\v\f\r",         // control whitespace
	string(rune(0x85)),   // unicode whitespace
	string(rune(0xA0)),   // unicode whitespace
	`"'`,                 // quotes
	"`",                  // backtick
	`\`,                  // backslash
}, "")

func isValidLongName(long string) bool {
	var (
		isEmpty     = long == ""
		hasBadStuff = strings.ContainsAny(long, badStuff)
		isValid     = !isEmpty && !hasBadStuff
	)
	return isValid
}

func getNameString(f Flag) string {
	var names []string
	if short, ok := f.GetShortName(); ok {
		names = append(names, "-"+string(short))
	}
	if long, ok := f.GetLongName(); ok {
		names = append(names, "--"+long)
	}
	return strings.Join(names, ", ")
}
