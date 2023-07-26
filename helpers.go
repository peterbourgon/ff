package ff

import (
	"fmt"
	"strconv"
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

func getPlaceholderFor(value any, usage string) string {
	// If the usage text contains a `backticked` substring, use that.
	for i := 0; i < len(usage); i++ {
		if usage[i] == '`' {
			for j := i + 1; j < len(usage); j++ {
				if usage[j] == '`' {
					return usage[i+1 : j]
				}
			}
			break
		}
	}

	// If the flag value implements a Placeholder method, use that.
	if p, ok := value.(interface{ Placeholder() string }); ok {
		if s := p.Placeholder(); s != "" {
			return s
		}
	}

	// Bool flags with default value false should have empty placeholders.
	if bf, ok := value.(interface{ IsBoolFlag() bool }); ok && bf.IsBoolFlag() {
		if b, err := strconv.ParseBool(fmt.Sprint(value)); err == nil && !b {
			return ""
		}
	}

	// Otherwise, use a transformation of the flag value type name.
	var typeName string
	typeName = fmt.Sprintf("%T", value)
	typeName = strings.ToUpper(typeName)
	if lastDot := strings.LastIndex(typeName, "."); lastDot > 0 {
		typeName = typeName[lastDot+1:]
	}
	typeName = strings.TrimSuffix(typeName, "VALUE")
	return typeName
}
