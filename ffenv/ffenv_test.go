package ffenv_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffenv"
	"github.com/peterbourgon/ff/v4/fftest"
)

func TestEnvFileParser(t *testing.T) {
	t.Parallel()

	testcases := fftest.TestCases{
		{
			ConfigFile: "testdata/empty.env",
			Want:       fftest.Vars{},
		},
		{
			ConfigFile:   "testdata/basic.env",
			Constructors: []fftest.Constructor{fftest.CoreConstructor},
			Want:         fftest.Vars{S: "bar", I: 99, B: true, D: time.Hour},
		},
		{
			ConfigFile:   "testdata/prefix.env",
			Constructors: []fftest.Constructor{fftest.CoreConstructor},
			Options:      []ff.Option{ff.WithEnvVarPrefix("MYPROG")},
			Want:         fftest.Vars{S: "bingo", I: 123},
		},
		{
			ConfigFile:   "testdata/prefix-undef.env",
			Constructors: []fftest.Constructor{fftest.CoreConstructor},
			Options:      []ff.Option{ff.WithEnvVarPrefix("MYPROG"), ff.WithConfigIgnoreUndefinedFlags()},
			Want:         fftest.Vars{S: "bango", I: 9},
		},
		{
			ConfigFile:   "testdata/quotes.env",
			Constructors: []fftest.Constructor{fftest.CoreConstructor},
			Want:         fftest.Vars{S: "", I: 32, X: []string{"1", "2 2", "3 3 3"}},
		},
		{
			ConfigFile:   "testdata/no-value.env",
			Constructors: []fftest.Constructor{fftest.CoreConstructor},
			Want:         fftest.Vars{WantParseErrorString: "DUR: parse error"},
		},
		{
			ConfigFile:   "testdata/spaces.env",
			Constructors: []fftest.Constructor{fftest.CoreConstructor},
			Want:         fftest.Vars{X: []string{"1", "2", "3", "4", "5", " 6", " 7 ", " 8 ", "9"}},
		},
		{
			ConfigFile:   "testdata/newlines.env",
			Constructors: []fftest.Constructor{fftest.CoreConstructor},
			Want:         fftest.Vars{S: "one\ntwo\nthree\n\n", X: []string{`A\nB\n\n`}},
		},
		{
			ConfigFile:   "testdata/comments.env",
			Constructors: []fftest.Constructor{fftest.CoreConstructor},
			Want:         fftest.Vars{S: "abc # def"},
		},
		{
			ConfigFile:   "testdata/short.env",
			Constructors: fftest.DefaultConstructors,
			Options:      []ff.Option{ff.WithEnvVarShortNames()},
			Want:         fftest.Vars{S: "hello", I: 99, D: 8 * time.Millisecond},
		},
		{
			ConfigFile:   "testdata/case-sensitive.env",
			Constructors: []fftest.Constructor{fftest.CoreConstructor},
			Options:      []ff.Option{ff.WithEnvVarPrefix("MYPREFIX"), ff.WithEnvVarCaseSensitive(), ff.WithConfigIgnoreUndefinedFlags()},
			Want:         fftest.Vars{S: "hello", I: 12345, D: 1*time.Minute + 30*time.Second},
		},
	}

	for i := range testcases {
		testcases[i].Constructors = []fftest.Constructor{
			fftest.CoreConstructor,
		}
		if testcases[i].Name == "" {
			testcases[i].Name = filepath.Base(testcases[i].ConfigFile)
		}
	}

	testcases.Run(t)
}

func TestAmbiguous(t *testing.T) {
	t.Parallel()

	fs := ff.NewFlagSet(t.Name())
	verboseFlag := fs.Bool('v', "verbose", "verbose output")
	versionFlag := fs.Bool('V', "version", "print version")

	for _, tc := range []struct {
		name        string
		options     []ff.Option
		wantErr     bool
		wantVerbose bool
		wantVersion bool
	}{
		{
			// With just the config file parser, env vars aren't enabled, so we
			// don't do duplicate detection up-front. And because we don't
			// provide WithEnvVarShortNames(), we're only using long names to
			// populate env2flags, so there is no v/V ambiguity. Also important:
			// the `V=true` doesn't match to anything in the env, but it *does*
			// match to a (short) flag name, because we didn't provide
			// WithConfigIgnoreFlagNames().
			name:        "ambiguous-1.env WithConfigFile",
			options:     []ff.Option{ff.WithConfigFile("testdata/ambiguous-1.env")},
			wantErr:     false,
			wantVerbose: true,
			wantVersion: true, // V=true should match to `-V, --version`
		},
		{
			// This is the same as above, but WithConfigIgnoreFlagNames() means
			// the `V=true` doesn't match to anything any more, and because we
			// didn't provide WithConfigIgnoreUndefinedFlags() that's an error.
			name:    "ambiguous-1.env WithConfigIgnoreFlagNames",
			options: []ff.Option{ff.WithConfigFile("testdata/ambiguous-1.env"), ff.WithConfigIgnoreFlagNames()},
			wantErr: true,
		},
		{
			// Same as above, but passing WithConfigIgnoreUndefinedFlags() means
			// it can now ignore `V=true` and succeed.
			name:        "ambiguous-1.env WithConfigIgnoreFlagNames",
			options:     []ff.Option{ff.WithConfigFile("testdata/ambiguous-1.env"), ff.WithConfigIgnoreFlagNames(), ff.WithConfigIgnoreUndefinedFlags()},
			wantErr:     false,
			wantVerbose: true,
			wantVersion: false,
		},
		{
			// WithEnvVarShortNames() by itself doesn't enable env var parsing,
			// and so doesn't trigger duplicate detection.
			name:        "ambiguous-1.env WithEnvVarShortNames",
			options:     []ff.Option{ff.WithConfigFile("testdata/ambiguous-1.env"), ff.WithEnvVarShortNames()},
			wantErr:     false,
			wantVerbose: true,
			wantVersion: true,
		},
		{
			// WithEnvVarShortNames() combined with WithEnvVars() triggers
			// duplicate detection up-front. Without WithEnvVarCaseSensitive(),
			// `-v` and `-V` both map to `V` and so are ambiguous, which results
			// in an error.
			name:    "ambiguous-1.env WithEnvVarShortNames WithEnvVars",
			options: []ff.Option{ff.WithConfigFile("testdata/ambiguous-1.env"), ff.WithEnvVarShortNames(), ff.WithEnvVars()},
			wantErr: true,
		},
		{
			// Same as above, but passing WithEnvVarCaseSensitive() means that
			// `-v` and `-V` are no longer ambiguous, and that `V=true` now
			// matches to `-V, --version`. But it also means that `VERSION=true`
			// is invalid, since it would need to be `version=true`. So, another
			// error.
			name:    "ambiguous-1.env WithEnvVarShortNames WithEnvVarCaseSensitive",
			options: []ff.Option{ff.WithConfigFile("testdata/ambiguous-1.env"), ff.WithEnvVarShortNames(), ff.WithEnvVarCaseSensitive()},
			wantErr: true,
		},
		{
			// But if we ignore undefined flags, then the parse can succeed.
			name:        "ambiguous-1.env WithEnvVarShortNames WithEnvVarCaseSensitive WithConfigIgnoreUndefinedFlags",
			options:     []ff.Option{ff.WithConfigFile("testdata/ambiguous-1.env"), ff.WithEnvVarShortNames(), ff.WithEnvVarCaseSensitive(), ff.WithConfigIgnoreUndefinedFlags()},
			wantErr:     false,
			wantVerbose: false,
			wantVersion: true,
		},
		{
			name:        "ambiguous-2.env matches both short names",
			options:     []ff.Option{ff.WithConfigFile("testdata/ambiguous-2.env")},
			wantErr:     false,
			wantVerbose: true,
			wantVersion: true,
		},
		{
			name:    "ambiguous-2.env WithConfigIgnoreFlagNames no match",
			options: []ff.Option{ff.WithConfigFile("testdata/ambiguous-2.env"), ff.WithConfigIgnoreFlagNames()},
			wantErr: true,
		},
		{
			name:    "ambiguous-2.env WithConfigIgnoreFlagNames WithEnvVars no match",
			options: []ff.Option{ff.WithConfigFile("testdata/ambiguous-2.env"), ff.WithConfigIgnoreFlagNames(), ff.WithEnvVars()},
			wantErr: true,
		},
		{
			name:    "ambiguous-2.env WithConfigIgnoreFlagNames WithEnvVars WithEnvVarShortNames duplicate",
			options: []ff.Option{ff.WithConfigFile("testdata/ambiguous-2.env"), ff.WithConfigIgnoreFlagNames(), ff.WithEnvVars(), ff.WithEnvVarShortNames()},
			wantErr: true,
		},
		{
			name:    "ambiguous-2.env WithConfigIgnoreFlagNames WithEnvVarShortNames ambiguous",
			options: []ff.Option{ff.WithConfigFile("testdata/ambiguous-2.env"), ff.WithConfigIgnoreFlagNames(), ff.WithEnvVarShortNames()},
			wantErr: true,
		},
		{
			name:        "ambiguous-2.env WithConfigIgnoreFlagNames WithEnvVarShortNames WithEnvVarCaseSensitive OK",
			options:     []ff.Option{ff.WithConfigFile("testdata/ambiguous-2.env"), ff.WithConfigIgnoreFlagNames(), ff.WithEnvVarShortNames(), ff.WithEnvVarCaseSensitive()},
			wantErr:     false,
			wantVerbose: true,
			wantVersion: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fs.Reset()

			options := append([]ff.Option{ff.WithConfigFileParser(ffenv.Parse)}, tc.options...)

			err := ff.Parse(fs, []string{}, options...)
			t.Logf("--verbose=%v --version=%v error=%v", *verboseFlag, *versionFlag, err)

			switch {
			case tc.wantErr:
				if want, have := tc.wantErr, err != nil; want != have {
					t.Errorf("error: want %v, have %v", want, have)
				}

			default:
				if want, have := tc.wantVerbose, *verboseFlag; want != have {
					t.Errorf("verbose: want %v, have %v", want, have)
				}
				if want, have := tc.wantVersion, *versionFlag; want != have {
					t.Errorf("version: want %v, have %v", want, have)
				}
			}
		})
	}
}
