package ff

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestFlagSetParse(t *testing.T) {
	var (
		s string
		d time.Duration
	)
	reset := func() {
		s = ""
		d = time.Duration(0)
	}
	define := func(fs *FlagSet) {
		fs.StringVar(&s, "s", "s default", "s usage")
		fs.DurationVar(&d, "d", time.Second, "d usage")
	}
	for _, testcase := range []struct {
		name      string
		args      []string
		json      string
		envPrefix string
		env       map[string]string
		wantStr   string
		wantDur   time.Duration
	}{
		{name: "empty",
			args:      []string{},
			json:      "",
			envPrefix: "", env: map[string]string{},
			wantStr: "s default",
			wantDur: time.Second,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			// Set up the JSON file, if any.
			var filename string
			if testcase.json != "" {
				f, _ := ioutil.TempFile(os.TempDir(), "ff_flag_set_test_")
				filename = f.Name()
				fmt.Fprintln(f, testcase.json)
				f.Close()
			}

			// Set up the environment, if any.
			for k, v := range testcase.env {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			// Define the flag set and parse.
			reset()
			fs := NewFlagSet(testcase.name)
			define(fs)
			if err := fs.Parse(testcase.args, FromEnvironment(testcase.envPrefix), FromJSONFile(filename)); err != nil {
				t.Fatalf("Parse: %v", err)
			}

			// Check.
			if want, have := testcase.wantStr, s; want != have {
				t.Errorf("s: want %q, have %q", want, have)
			}
			if want, have := testcase.wantDur, d; want != have {
				t.Errorf("d: want %v, have %v", want, have)
			}
		})
	}

}
