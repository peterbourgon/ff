package ffhelp_test

import (
	"strings"
	"testing"

	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffhelp"
	"github.com/peterbourgon/ff/v4/fftest"
)

func TestFlagsHelp(t *testing.T) {
	t.Parallel()

	t.Run("basic", func(t *testing.T) {
		fs := ff.NewFlags("fftest")
		fs.Duration('d', "dur", 0, "duration flag")
		fs.String('s', "str", "", "string flag")

		want := strings.TrimSpace(testFlagsHelpBasic)
		have := strings.TrimSpace(ffhelp.FlagsHelp(fs))
		if want != have {
			t.Errorf("\n%s", fftest.DiffString(want, have))
		}
	})

	t.Run("summary", func(t *testing.T) {
		fs := ff.NewFlags("fftest")
		fs.Duration('d', "dur", 0, "duration flag")
		fs.String('s', "str", "", "string flag")

		want := strings.TrimSpace(testFlagsHelpSummary)
		have := strings.TrimSpace(ffhelp.FlagsHelpSummary(fs, "do something interesting"))
		if want != have {
			t.Errorf("\n%s", fftest.DiffString(want, have))
		}
	})

	t.Run("details", func(t *testing.T) {
		fs := ff.NewFlags("fftest")
		fs.Duration('d', "dur", 0, "duration flag")
		fs.String('s', "str", "", "string flag")

		want := strings.TrimSpace(testFlagsHelpDetails)
		have := strings.TrimSpace(ffhelp.NewFlagsSections(fs, "do something interesting", strings.TrimSpace(loremIpsum)).String())
		if want != have {
			t.Errorf("\n%s", fftest.DiffString(want, have))
		}
	})
}

var testFlagsHelpBasic = `
fftest

FLAGS
  -d, --dur DURATION   duration flag (default: 0s)
  -s, --str STRING     string flag
`

var testFlagsHelpSummary = `
fftest -- do something interesting

FLAGS
  -d, --dur DURATION   duration flag (default: 0s)
  -s, --str STRING     string flag
`

var testFlagsHelpDetails = `
fftest -- do something interesting

Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nam diam eros,
vestibulum at pulvinar vulputate, vehicula id lacus. Class aptent taciti
sociosqu ad litora torquent per conubia nostra, per inceptos himenaeos.
Mauris venenatis felis orci, ac consectetur mi molestie ac. Integer pharetra
pharetra odio. Maecenas metus eros, viverra eget efficitur ut, feugiat in
tortor. Quisque elit nibh, rhoncus in posuere et, bibendum non turpis.
Maecenas eget dui malesuada, pretium tellus quis, bibendum felis. Duis erat
enim, faucibus id auctor ac, ornare sed metus.

FLAGS
  -d, --dur DURATION   duration flag (default: 0s)
  -s, --str STRING     string flag
`

func TestFlagsHelp_OnlyLong(t *testing.T) {
	t.Parallel()

	fs := ff.NewFlags("fftest")
	fs.BoolLong("alpha", false, "alpha usage")
	fs.BoolLong("beta", false, "beta usage")

	want := strings.TrimSpace(testFlagsHelpOnlyLong)
	have := strings.TrimSpace(ffhelp.FlagsHelp(fs))
	if want != have {
		t.Errorf("\n%s", fftest.DiffString(want, have))
	}
}

const testFlagsHelpOnlyLong = `
fftest

FLAGS
  --alpha   alpha usage (default: false)
  --beta    beta usage (default: false)
`
