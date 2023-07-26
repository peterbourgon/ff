package ff

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

const defaultUsageIndent = "  "

// DefaultFlagsUsage produces a simple usage text for the flag set. It's meant
// to be a reasonable default, and to serve as an example. Users with more
// sophisticated needs should implement their own usage function.
func DefaultFlagsUsage(fs Flags) string {
	var buf bytes.Buffer

	fmt.Fprintf(&buf, "COMMAND\n")
	fmt.Fprintf(&buf, "%s%s\n", defaultUsageIndent, fs.GetName())
	fmt.Fprintf(&buf, "\n")

	writeFlagGroups(fs, &buf)

	return strings.TrimSpace(buf.String())
}

// DefaultCommandUsage produces a simple usage text for the command. It's meant
// to be a reasonable default, and to serve as an example. Users with more
// sophisticated needs should implement their own usage function.
func DefaultCommandUsage(cmd *Command) string {
	cmd = cmd.GetSelected()
	if cmd == nil {
		return ""
	}

	var buf bytes.Buffer

	{
		name := cmd.Name
		if cmd.ShortHelp != "" {
			name = fmt.Sprintf("%s -- %s", name, cmd.ShortHelp)
		}
		fmt.Fprintf(&buf, "COMMAND\n")
		fmt.Fprintf(&buf, "%s%s\n", defaultUsageIndent, name)
		fmt.Fprintf(&buf, "\n")
	}

	{
		if cmd.Usage != "" {
			fmt.Fprintf(&buf, "USAGE\n")
			fmt.Fprintf(&buf, "%s%s\n", defaultUsageIndent, cmd.Usage)
			fmt.Fprintf(&buf, "\n")
		}
	}

	{
		if cmd.LongHelp != "" {
			fmt.Fprintf(&buf, "%s\n", cmd.LongHelp)
			fmt.Fprintf(&buf, "\n")
		}
	}

	{
		if len(cmd.Subcommands) > 0 {
			fmt.Fprintf(&buf, "SUBCOMMANDS\n")
			tw := newTabWriter(&buf)
			for _, sc := range cmd.Subcommands {
				fmt.Fprintf(tw, "%s%s\t%s\n", defaultUsageIndent, sc.Name, sc.ShortHelp)
			}
			tw.Flush()
			fmt.Fprintf(&buf, "\n")
		}
	}

	{
		writeFlagGroups(cmd.Flags, &buf)
	}

	return strings.TrimSpace(buf.String())
}

//
//
//

func writeFlagGroups(fs Flags, w io.Writer) {
	for i, g := range makeFlagGroups(fs) {
		switch {
		case i == 0, g.name == "":
			fmt.Fprintf(w, "FLAGS\n")
		default:
			fmt.Fprintf(w, "FLAGS (%s)\n", g.name)
		}
		for _, line := range g.help {
			fmt.Fprintf(w, "%s%s\n", defaultUsageIndent, line)
		}
		fmt.Fprintf(w, "\n")
	}
}

type flagGroup struct {
	name string
	help []string
}

func makeFlagGroups(fs Flags) []flagGroup {
	var (
		order = []string{}
		index = map[string][]Flag{}
	)
	fs.WalkFlags(func(f Flag) error {
		name := f.GetFlagsName()
		if _, ok := index[name]; !ok {
			order = append(order, name)
		}
		index[name] = append(index[name], f)
		return nil
	})

	groups := make([]flagGroup, 0, len(order))
	for _, name := range order {
		flags := index[name]
		help := getFlagsHelp(flags)
		groups = append(groups, flagGroup{
			name: name,
			help: help,
		})
	}

	return groups
}

func getFlagsHelp(flags []Flag) []string {
	var haveShortFlags bool
	for _, f := range flags {
		if _, ok := f.GetShortName(); ok {
			haveShortFlags = true
			break
		}
	}

	var buf bytes.Buffer

	tw := newTabWriter(&buf)
	for _, f := range flags {
		var (
			short, hasShort = f.GetShortName()
			long, hasLong   = f.GetLongName()
			cf, isCoreFlag  = f.(*coreFlag)
			longAsShort     = !hasShort && hasLong && isCoreFlag && cf.flagSet.isStdAdapter
			flagNames       string
		)
		switch {
		case longAsShort:
			flagNames = fmt.Sprintf("-%s", long)
		case !longAsShort && hasShort && hasLong:
			flagNames = fmt.Sprintf("-%s, --%s", string(short), long)
		case !longAsShort && hasShort && !hasLong:
			flagNames = fmt.Sprintf("-%s", string(short))
		case !longAsShort && !hasShort && hasLong && haveShortFlags:
			flagNames = fmt.Sprintf("    --%s", long)
		case !longAsShort && !hasShort && hasLong && !haveShortFlags:
			flagNames = fmt.Sprintf("--%s", long)
		}

		if p := f.GetPlaceholder(); p != "" {
			flagNames = fmt.Sprintf("%s %s", flagNames, p)
		}

		flagUsage := f.GetUsage()
		if d := f.GetDefault(); d != "" {
			flagUsage = fmt.Sprintf("%s (default: %s)", flagUsage, d)
		}

		fmt.Fprintf(tw, "%s\t%s\n", flagNames, flagUsage)
	}
	tw.Flush()

	return strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
}

func newTabWriter(w io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
}
