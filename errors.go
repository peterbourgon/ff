package ff

import (
	"errors"
	"flag"
)

var (
	// ErrHelp should be returned by the parse method of flag sets, when the
	// provided args indicate the user has requested help. Usually this means
	// -h or --help was one of the args.
	ErrHelp = flag.ErrHelp

	// ErrDuplicateFlag is returned by the core flag set, if a flag is added
	// that has the same name as an existing flag.
	ErrDuplicateFlag = errors.New("duplicate flag")

	// ErrNotParsed may be returned by flag sets, when a method is called that
	// requires the flag set to have been successfully parsed, but it hasn't
	// been.
	ErrNotParsed = errors.New("not parsed")

	// ErrAlreadyParsed may be returned by the parse method of flag sets, if the
	// flag set has already been successfully parsed, and cannot be parsed
	// again.
	ErrAlreadyParsed = errors.New("already parsed")

	// ErrNoExec is returned when a command without an exec function is run.
	ErrNoExec = errors.New("no exec function")

	// ErrUnknownFlag should be returned by flag sets, when a specific or
	// user-requested flag could not be found.
	ErrUnknownFlag = errors.New("unknown flag")
)
