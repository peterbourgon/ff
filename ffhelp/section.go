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

// NewSection returns a section with the given title and lines.
func NewSection(title string, lines ...string) *Section {
	return &Section{
		Title:      title,
		Lines:      lines,
		LinePrefix: DefaultLinePrefix,
	}
}

// NewUntitledSection returns a section with only the given lines, and no
// title.
func NewUntitledSection(lines ...string) *Section {
	return &Section{
		Lines: lines,
	}
}

// NewFlagsSection returns a section with the title "FLAGS", and one line for
// every flag defined in the provided set of flags, via [FlagSpec].
func NewFlagsSection(fs ff.Flags) *Section {
	sections := makeFlagSections(flagSectionsConfig{
		Flags:         fs,
		SingleSection: true,
	})
	if len(sections) != 1 {
		panic(fmt.Errorf("expected 1 section, got %d", len(sections)))
	}
	return sections[0]
}

// NewSubcommandsSection returns a section with the title "SUBCOMMANDS", and one
// line for every subcommand in the slice. Lines consist of the subcommand name
// and the ShortHelp for that subcommand, in a tab-delimited columnar format.
func NewSubcommandsSection(subcommands []*ff.Command) *Section {
	var lines []string
	for _, sc := range subcommands {
		lines = append(lines, fmt.Sprintf("%s\t%s\n", sc.Name, sc.ShortHelp))
	}
	if len(lines) <= 0 {
		lines = append(lines, "(no subcommands)")
	}
	return &Section{
		Title:       "SUBCOMMANDS",
		Lines:       lines,
		LinePrefix:  DefaultLinePrefix,
		LineColumns: true,
	}
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

//
//
//

// Sections is an ordered list of [Section] values, rendered together and each
// separated by a single blank line.
type Sections []*Section

// NewFlagsSections returns a list of sections describing the provided flags.
//
// The first section is an untitled section containing [ff.Flags.GetName]. If
// summary is non-empty, it will be included as a suffix, after two hyphens.
//
// If detail string(s) are provided, the next section will be another untitled
// section, containing only those detail strings as the lines.
//
// The final section(s) are titled "FLAGS", or "FLAGS (name)" in the case of
// parent flag sets. Every unique flag set observed by a call to [ff.Flags.Walk]
// will correspond to a separate flags section in the returned value.
func NewFlagsSections(fs ff.Flags, summary string, details ...string) Sections {
	var sections Sections

	name := fs.GetName()
	if summary != "" {
		name = fmt.Sprintf("%s -- %s", name, summary)
	}
	sections = append(sections, NewUntitledSection(name))

	if len(details) > 0 {
		sections = append(sections, NewUntitledSection(details...))
	}

	sections = append(sections, makeFlagSections(flagSectionsConfig{
		Flags:           fs,
		SharedAlignment: true,
	})...)

	return sections
}

// NewCommandSections returns a slice of sections describing the given command.
func NewCommandSections(cmd *ff.Command) Sections {
	var sections Sections

	if selected := cmd.GetSelected(); selected != nil {
		cmd = selected
	}

	commandTitle := cmd.Name
	if cmd.ShortHelp != "" {
		commandTitle = fmt.Sprintf("%s -- %s", commandTitle, cmd.ShortHelp)
	}
	sections = append(sections, NewUntitledSection(commandTitle))

	if cmd.Usage != "" {
		sections = append(sections, NewSection("USAGE", cmd.Usage))
	}

	if cmd.LongHelp != "" {
		sections = append(sections, &Section{Lines: []string{cmd.LongHelp}})
	}

	if len(cmd.Subcommands) > 0 {
		sections = append(sections, NewSubcommandsSection(cmd.Subcommands))
	}

	sections = append(sections, makeFlagSections(flagSectionsConfig{
		Flags:           cmd.Flags,
		SharedAlignment: true,
	})...)

	return sections
}

// WriteTo implements [io.WriterTo].
func (ss Sections) WriteTo(w io.Writer) (n int64, _ error) {
	if len(ss) <= 0 {
		return 0, nil
	}

	for i, s := range ss {
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
func (ss Sections) String() string {
	var buf bytes.Buffer
	if _, err := ss.WriteTo(&buf); err != nil {
		return fmt.Sprintf("%%!ERROR<%v>", err)
	}
	return buf.String()
}

//
//
//

type flagSectionsConfig struct {
	Flags           ff.Flags
	SingleSection   bool // treat all parent flags as belonging to the base flag set
	AlwaysSubtitle  bool // append fs.GetName() to every title
	SharedAlignment bool // use the same column spacing across all sections
}

// makeFlagSections produces FLAGS sections (only) for a flag set.
func makeFlagSections(cfg flagSectionsConfig) Sections {
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
		sections = Sections{}
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

	return sections
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
