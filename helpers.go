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

func isValidLongName(long string) bool {
	return long != ""
}

func getNameStrings(f Flag) []string {
	var names []string
	if short, ok := f.GetShortName(); ok {
		names = append(names, string(short))
	}
	if long, ok := f.GetLongName(); ok {
		names = append(names, long)
	}
	return names
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
