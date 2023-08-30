package ffhelp

import (
	"bytes"
	"fmt"
	"io"

	"github.com/peterbourgon/ff/v4"
)

// Help represents help output for a flag set, command, etc.
type Help []Section

// Flags returns a default [Help] value representing fs. If details are
// provided, they're included as a single untitled section before any FLAGS
// section(s).
//
// This function is meant as reasonable default for most users, and as an
// example. Callers who want different help output should implement their own
// [Help] value constructors like this one.
func Flags(fs ff.Flags, details ...string) Help {
	var help Help

	top := fs.GetName()
	if d, ok := fs.(ff.Describer); ok {
		if desc := d.GetDescription(); desc != "" {
			top = desc
		}
	}
	help = append(help, NewUntitledSection(top))

	if len(details) > 0 {
		help = append(help, NewUntitledSection(details...))
	}

	help = append(help, NewFlagsSections(fs)...)

	return help
}

// Command returns [Help] for the given command.
//
// This function is meant as reasonable default for most users, and as an
// example. Callers who want different help output should implement their own
// [Help] value constructors like this one.
func Command(cmd *ff.Command) Help {
	var help Help

	if selected := cmd.GetSelected(); selected != nil {
		cmd = selected
	}

	commandTitle := cmd.Name
	if cmd.ShortHelp != "" {
		commandTitle = fmt.Sprintf("%s -- %s", commandTitle, cmd.ShortHelp)
	}
	help = append(help, NewUntitledSection(commandTitle))

	if cmd.Usage != "" {
		help = append(help, NewSection("USAGE", cmd.Usage))
	}

	if cmd.LongHelp != "" {
		help = append(help, NewUntitledSection(cmd.LongHelp))
	}

	if len(cmd.Subcommands) > 0 {
		help = append(help, NewSubcommandsSection(cmd.Subcommands))
	}

	help = append(help, NewFlagsSections(cmd.Flags)...)

	return help
}

// WriteTo implements [io.WriterTo].
func (h Help) WriteTo(w io.Writer) (n int64, _ error) {
	if len(h) <= 0 {
		return 0, nil
	}

	for i, s := range h {
		if i > 0 {
			nn, err := fmt.Fprintf(w, "\n")
			if err != nil {
				return n, err
			}
			n += int64(nn)
		}

		nn, err := s.WriteTo(w) // always ends in \n
		if err != nil {
			return n, err
		}
		n += int64(nn)
	}

	return n, nil
}

// String implements [fmt.Stringer].
func (h Help) String() string {
	var buf bytes.Buffer
	if _, err := h.WriteTo(&buf); err != nil {
		return fmt.Sprintf("%%!ERROR<%v>", err)
	}
	return buf.String()
}
