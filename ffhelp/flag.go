package ffhelp

import (
	"fmt"
	"io"
	"strings"

	"github.com/peterbourgon/ff/v4"
)

// Flag wraps [ff.Flag] to implement [fmt.Formatter]. It's relatively low-level,
// most users should prefer higher-level helper functions.
type Flag struct{ ff.Flag }

// WrapFlag wraps the provided [ff.Flag] with [Flag], allowing it to be rendered
// as a string via [Flag.Format].
func WrapFlag(f ff.Flag) Flag { return Flag{Flag: f} }

// FormatFlag wraps the provided [ff.Flag] with a [Flag], and formats the flag
// according to the provided format string.
func FormatFlag(f ff.Flag, format string) string { return fmt.Sprintf(format, Flag{Flag: f}) }

// Format implements [fmt.Formatter] with support for the following verbs.
//
//	VERB   DESCRIPTION                              EXAMPLE
//	%s     short and long name, comma delimited     "f, foo"
//	%+s    like %s with hyphen prefixes             "-f, --foo"
//	%#+s   like %+s with empty short names padded   "    --foo"
//	%v     like %s with placeholder suffix          "f, foo STR"
//	%+v    like %+s with placeholder suffix         "-f, --foo STR"
//	%#+v   like %#+s with placeholder suffix        "    --foo STR"
//	%n     short name                               "f"
//	%+n    short name with one-hyphen prefix        "-f"
//	%l     long name                                "foo"
//	%+l    long name with two-hyphen prefix         "--foo"
//	%u     usage text                               "foo parameter"
//	%k     placeholder                              "STR"
//	%d     default value                            "bar"
//
// See the tests for more complete examples.
func (f Flag) Format(s fmt.State, verb rune) {
	if f.Flag == nil {
		fmt.Fprintf(s, "%%!%c<nil>", verb)
		return
	}

	switch verb {
	case 's', 'v', 'x':
		addHyphens := s.Flag('+')
		addPadding := addHyphens && s.Flag('#')
		short, haveShort := f.GetShortName()
		long, haveLong := f.GetLongName()

		var shortstr string
		if haveShort {
			switch {
			case addHyphens:
				shortstr = "-" + string(short)
			case !addHyphens:
				shortstr = string(short)
			}
		}

		if haveLong && addHyphens {
			long = "--" + long
		}

		switch {
		case haveShort && haveLong:
			fmt.Fprintf(s, "%s, %s", shortstr, long)
		case haveShort && !haveLong:
			fmt.Fprintf(s, "%s", shortstr)
		case !haveShort && haveLong && addPadding:
			fmt.Fprintf(s, "    %s", long)
		case !haveShort && haveLong && !addPadding:
			fmt.Fprintf(s, "%s", long)
		}

		if verb == 'v' {
			if p := f.GetPlaceholder(); p != "" {
				fmt.Fprintf(s, " %s", p)
			}
		}

	case 'n':
		short, ok := f.GetShortName()
		switch {
		case !ok:
			//
		case s.Flag('+'):
			io.WriteString(s, "-"+string(short))
		default:
			io.WriteString(s, string(short))
		}

	case 'l':
		long, ok := f.GetLongName()
		switch {
		case !ok:
			//
		case s.Flag('+'):
			io.WriteString(s, "--"+long)
		default:
			io.WriteString(s, long)
		}

	case 'd':
		io.WriteString(s, f.GetDefault())

	case 'u':
		io.WriteString(s, f.GetUsage())

	case 'k':
		io.WriteString(s, f.GetPlaceholder())
	}
}

//
//
//

// FlagSpec represents a single-line help text for an [ff.Flag]. That line
// consists of two parts: args, which is a fixed-width formatted description of
// the flag names and placeholder; and help, which is a combination of the usage
// string and the default value (if non-empty).
type FlagSpec struct {
	Flag ff.Flag
	Args string // "-f, --foo STR"
	Help string // "value of foo parameter (default: bar)"
}

// MakeFlagSpec produces a [FlagSpec] from an [ff.Flag].
func MakeFlagSpec(f ff.Flag) FlagSpec {
	ff := Flag{f}

	args := fmt.Sprintf("%#+v", ff)
	if sf, ok := f.(interface{ IsStdFlag() bool }); ok && sf.IsStdFlag() {
		args = strings.Replace(args, "--", "-", 1)
		args = strings.TrimSpace(args)
	}

	help := fmt.Sprintf("%u", ff)
	if d := f.GetDefault(); d != "" {
		help = fmt.Sprintf("%s (default: %s)", help, d)
	}

	return FlagSpec{
		Flag: f,
		Args: args,
		Help: help,
	}
}

// String returns a tab-delimited and newline-terminated string containing args
// and help in that order. It's intended to be written to a [tabwriter.Writer].
func (fs FlagSpec) String() string {
	return fmt.Sprintf("%s\t%s\n", fs.Args, fs.Help)
}
