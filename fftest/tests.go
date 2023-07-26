package fftest

import (
	"errors"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/peterbourgon/ff/v4"
)

// TestFlagSet checks the core invariants of a flag set and its flags. The
// provided flag set should contain at least two flags, and calling parse with
// the provided args should succeed.
func TestFlagSet(t *testing.T, fs ff.FlagSet, args []string) {
	t.Helper()

	if fs.GetFlagSetName() == "" {
		t.Errorf("GetFlagSetName: empty")
	}

	if fs.IsParsed() {
		t.Errorf("IsParsed: initially true")
	}

	if args := fs.GetArgs(); len(args) != 0 {
		t.Errorf("GetArgs: initially non-empty (%v)", args)
	}

	var flags []ff.Flag
	if err := fs.WalkFlags(func(f ff.Flag) error {
		flags = append(flags, f)
		return nil
	}); err != nil {
		t.Fatalf("WalkFlags: error: %v", err)
	}
	if n := len(flags); n < 2 {
		t.Fatalf("WalkFlags: need at least 2 flags, have %d", n)
	}

	var count int
	errEarlyReturn := errors.New("early return")
	if err := fs.WalkFlags(func(f ff.Flag) error {
		count++
		return errEarlyReturn
	}); !errors.Is(err, errEarlyReturn) {
		t.Errorf("WalkFlags: received error (%v) not errors.Is with returned error (%v)", err, errEarlyReturn)
	}
	if count != 1 {
		t.Errorf("WalkFlags: should have walked 1 flag, have %d", count)
	}

	if _, ok := fs.GetFlag(""); ok {
		t.Errorf("GetFlag: passing an empty string returned ok=true")
	}

	for i, f := range flags {
		var (
			name             = getFlagName(f)
			short, haveShort = f.GetShortName()
			long, haveLong   = f.GetLongName()
		)

		if !haveShort && !haveLong {
			t.Errorf("flag (%d/%d) has neither short nor long name", i+1, len(flags))
			continue
		}

		if haveShort && (short == 0 || short == utf8.RuneError) {
			t.Errorf("%s: GetShortName: returned invalid rune (%x) with ok=true", name, short)
			haveShort = false
		}

		if haveLong && (long == "") {
			t.Errorf("%s: GetLongName: returned empty string with ok=true", name)
			haveLong = false
		}

		if f.IsSet() {
			t.Errorf("%s: IsSet: returned true before parse", name)
		}

		if f.GetFlagSetName() == "" {
			t.Errorf("%s: GetFlagSetName: returned empty string", name)
		}

		if f.GetDefault() != f.GetValue() {
			t.Errorf("%s: GetDefault (%q) != GetValue (%q) before being set", name, f.GetDefault(), f.GetValue())
		}

		if haveShort {
			if ff, ok := fs.GetFlag(string(short)); !ok {
				t.Errorf("%s: GetFlag(%s): returned ok=false", name, string(short))
			} else if ff != f {
				t.Errorf("%s: GetFlag(%s): returned different flag (%s)", name, string(short), getFlagName(ff))
			}
		}

		if haveLong {
			if ff, ok := fs.GetFlag(long); !ok {
				t.Errorf("%s: GetFlag(%s): returned ok=false", name, long)
			} else if ff != f {
				t.Errorf("%s: GetFlag(%s): returned different flag (%s)", name, long, getFlagName(ff))
			}
		}
	}

	if err := ff.Parse(fs, args); err != nil {
		t.Fatalf("Parse: error: %v", err)
	}

	if !fs.IsParsed() {
		t.Errorf("IsParsed: returned false after successful parse")
	}
}

func getFlagName(f ff.Flag) string {
	var names []string
	if short, ok := f.GetShortName(); ok {
		names = append(names, string(short))
	}
	if long, ok := f.GetLongName(); ok {
		names = append(names, long)
	}
	return strings.Join(names, ", ")
}
