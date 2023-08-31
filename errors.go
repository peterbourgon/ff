package ff

import (
	"errors"
	"flag"
)

var (
	// ErrHelp should be returned by flag sets during parse, when the provided
	// args indicate the user has requested help.
	ErrHelp = flag.ErrHelp

	// ErrDuplicateFlag should be returned by flag sets when the user tries to
	// add a flag with the same name as a pre-existing flag.
	ErrDuplicateFlag = errors.New("duplicate flag")

	// ErrNotParsed may be returned by flag set methods which require the flag
	// set to have been successfully parsed, and that condition isn't satisfied.
	ErrNotParsed = errors.New("not parsed")

	// ErrAlreadyParsed may be returned by the parse method of flag sets, if the
	// flag set has already been successfully parsed, and cannot be parsed
	// again.
	ErrAlreadyParsed = errors.New("already parsed")

	// ErrUnknownFlag should be returned by flag sets methods to indicate that a
	// specific or user-requested flag was provided but could not be found.
	ErrUnknownFlag = errors.New("unknown flag")

	// ErrNoExec is returned when a command without an exec function is run.
	ErrNoExec = errors.New("no exec function")
)
