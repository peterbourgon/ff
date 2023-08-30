package ffhelp

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/peterbourgon/ff/v4"
)

// DefaultLinePrefix is used by [Section] constructors in this package.
var DefaultLinePrefix = "  "

// Section describes a single block of help text. A section typically begins
// with a typically uppercase TITLE, and contains one or more lines of content,
// which are typically indented by LinePrefix.
type Section struct {
	// Title of the section, typically UPPERCASE.
	Title string

	// Lines in the section. Lines should be formatted for the user device, i.e.
	// long lines should be split into multiple lines. Each line is prefixed
	// with LinePrefix before being rendered.
	Lines []string

	// LinePrefix is prefixed to each line when the section is rendered.
	LinePrefix string

	// LineColumns indicates that each line is a tab-delimited set of fields,
	// and therefore will be rendered in a columnar format via text/tabwriter.
	LineColumns bool
}

// WriteTo implements [io.WriterTo], always ending with a newline.
func (s Section) WriteTo(w io.Writer) (n int64, _ error) {
	if s.Title != "" {
		nn, err := fmt.Fprint(w, ensureNewline(s.Title))
		if err != nil {
			return n, err
		}
		n += int64(nn)
	}

	dst, flush := w, func() error { return nil }
	if s.LineColumns {
		tab := newTabWriter(w)
		dst, flush = tab, tab.Flush
	}

	for _, line := range s.Lines {
		nn, err := fmt.Fprint(dst, ensureNewline(s.LinePrefix+line))
		if err != nil {
			return n, err
		}
		n += int64(nn)
	}
	if err := flush(); err != nil {
		return n, err
	}

	return n, nil
}

// String returns a multi-line string representation ending with a newline.
func (s Section) String() string {
	var buf bytes.Buffer
	if _, err := s.WriteTo(&buf); err != nil {
		return fmt.Sprintf("%%!ERROR<%v>", err)
	}
	return buf.String()
}

// NewSection returns a section with the given title and lines.
func NewSection(title string, lines ...string) Section {
	return Section{
		Title:      title,
		Lines:      lines,
		LinePrefix: DefaultLinePrefix,
	}
}

// NewUntitledSection returns a section with no title and the given lines.
func NewUntitledSection(lines ...string) Section {
	return Section{
		Lines: lines,
	}
}

// NewFlagsSection produces a single FLAGS section representing every flag
// available to fs. Each flag is rendered via [FlagSpec].
func NewFlagsSection(fs ff.Flags) Section {
	ss := newFlagSections(flagSectionsConfig{Flags: fs, SingleSection: true})
	if len(ss) != 1 {
		panic(fmt.Errorf("expected 1 section, got %d", len(ss)))
	}
	return ss[0]
}

// NewFlagsSections returns FLAG section(s) representing every flag available to
// fs. Flags are grouped into sections according to their parent flag set. Each
// flag is rendered via [FlagSpec].
func NewFlagsSections(fs ff.Flags) []Section {
	return newFlagSections(flagSectionsConfig{Flags: fs, SharedAlignment: true})
}

// NewSubcommandsSection returns a section with the title "SUBCOMMANDS", and one
// line for every subcommand in the slice. Lines consist of the subcommand name
// and the ShortHelp for that subcommand, in a tab-delimited columnar format.
func NewSubcommandsSection(subcommands []*ff.Command) Section {
	var lines []string
	for _, sc := range subcommands {
		lines = append(lines, fmt.Sprintf("%s\t%s\n", sc.Name, sc.ShortHelp))
	}
	if len(lines) <= 0 {
		lines = append(lines, "(no subcommands)")
	}
	return Section{
		Title:       "SUBCOMMANDS",
		Lines:       lines,
		LinePrefix:  DefaultLinePrefix,
		LineColumns: true,
	}
}

//
//
//

type flagSectionsConfig struct {
	Flags           ff.Flags
	SingleSection   bool // treat all flags as belonging to the base flag set
	AlwaysSubtitle  bool // add the flag set name to every section title
	SharedAlignment bool // use the same column spacing across all sections
}

func newFlagSections(cfg flagSectionsConfig) []Section {
	var (
		index = map[string][]ff.Flag{}
		order = []string{}
	)
	cfg.Flags.WalkFlags(func(f ff.Flag) error {
		var parent string
		if cfg.SingleSection {
			parent = cfg.Flags.GetName()
		} else {
			parent = f.GetFlags().GetName()
		}
		if _, ok := index[parent]; !ok {
			order = append(order, parent)
		}
		index[parent] = append(index[parent], f)
		return nil
	})

	var (
		buffer   = &bytes.Buffer{}
		tab      = newTabWriter(buffer)
		flushOne func() error
		flushAll func() error
	)
	if cfg.SharedAlignment {
		flushOne = func() error { return nil }
		flushAll = tab.Flush
	} else {
		flushOne = tab.Flush
		flushAll = func() error { return nil }
	}

	for _, name := range order {
		flags := index[name]
		if len(flags) <= 0 {
			continue
		}
		for _, f := range flags {
			fmt.Fprint(tab, MakeFlagSpec(f).String())
		}
		if err := flushOne(); err != nil {
			panic(err)
		}
	}
	if err := flushAll(); err != nil {
		panic(err)
	}

	var (
		lines    = splitLines(buffer.String())
		sections = []*Section{}
	)
	for i, name := range order {
		flags := index[name]
		if len(flags) <= 0 {
			continue
		}

		if len(lines) < len(flags) {
			panic(fmt.Errorf("%s: flag count %d, remaining section line count %d", name, len(flags), len(lines)))
		}

		sectionLines := lines[:len(flags)]
		if len(sectionLines) <= 0 {
			panic(fmt.Errorf("%s: flag count %d, section line count 0", name, len(flags)))
		}

		title := "FLAGS"
		if cfg.AlwaysSubtitle || i > 0 {
			title = fmt.Sprintf("%s (%s)", title, name)
		}

		sections = append(sections, &Section{
			Title:      title,
			Lines:      sectionLines,
			LinePrefix: DefaultLinePrefix,
		})

		lines = lines[len(flags):]
	}

	var (
		mindexOne = -1
		mindexAll = -1
	)
	for _, s := range sections {
		for _, line := range s.Lines {
			var index int
			for index < len(line) && line[index] == ' ' {
				index++
			}
			switch {
			case mindexOne < 0 || index < mindexOne:
				mindexOne = index
			case mindexAll < 0 || index < mindexAll:
				mindexAll = index
			}
		}
		if mindexOne > 0 && !cfg.SharedAlignment {
			for i := range s.Lines {
				s.Lines[i] = s.Lines[i][mindexOne:]
			}
		}
	}
	if mindexAll > 0 && cfg.SharedAlignment {
		for _, s := range sections {
			for i := range s.Lines {
				s.Lines[i] = s.Lines[i][mindexOne:]
			}
		}
	}

	flat := make([]Section, len(sections))
	for i := range sections {
		flat[i] = *sections[i]
	}

	return flat
}

//
//
//

func newTabWriter(w io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
}

func ensureNewline(s string) string {
	return strings.TrimSuffix(s, "\n") + "\n"
}

func splitLines(s string) []string {
	var res []string
	for _, line := range strings.Split(s, "\n") {
		if line == "" {
			continue
		}
		res = append(res, line)
	}
	return res
}
