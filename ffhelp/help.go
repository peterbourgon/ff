package ffhelp

import (
	"github.com/peterbourgon/ff/v4"
)

// FlagsHelp is a helper function which calls [NewFlagsSections] and returns a
// multi-line help string describing the provided flags.
func FlagsHelp(fs ff.Flags) string {
	return NewFlagsSections(fs, "").String()
}

// FlagsHelpSummary is a helper function which calls [NewFlagsSections] and
// returns a multi-line help string describing the provided flags. If the
// summary string is non-empty, it's included after the name of the flag set at
// the top of the help text.
func FlagsHelpSummary(fs ff.Flags, summary string) string {
	return NewFlagsSections(fs, summary).String()
}

// CommandHelp is a helper function which calls [NewCommandSections] and retrns
// a multi-line help string describing the provided command.
func CommandHelp(cmd *ff.Command) string {
	return NewCommandSections(cmd).String()
}
