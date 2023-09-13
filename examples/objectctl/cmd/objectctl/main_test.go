package main

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/examples/objectctl/pkg/objectapi"
	"github.com/peterbourgon/ff/v4/fftest"
)

func TestExec(t *testing.T) {
	t.Parallel()

	type testcase struct {
		name       string
		args       []string
		wantErr    error
		wantStdout string
		wantStderr string
	}

	testcases := []testcase{
		{
			name:       "no args",
			wantStderr: rootUsage,
			wantErr:    ff.ErrNoExec,
		},
		{
			name:       "-h",
			args:       []string{"-h"},
			wantStderr: rootUsage,
			wantErr:    ff.ErrHelp,
		},
		{
			name:       "list",
			args:       []string{"list"},
			wantStderr: listUsage,
			wantErr:    objectapi.ErrUnauthorized,
		},
		{
			name:       "list -h",
			args:       []string{"list", "-h"},
			wantStderr: listUsage,
			wantErr:    ff.ErrHelp,
		},
		{
			name:       "list --token=SECRET",
			args:       []string{"list", "--token=SECRET"},
			wantStdout: listOutput,
		},
	}

	for _, test := range testcases {
		t.Run(test.name, func(t *testing.T) {
			var (
				ctx     = context.Background()
				stdin   = strings.NewReader("")
				stdout  = &bytes.Buffer{}
				stderr  = &bytes.Buffer{}
				haveErr = exec(ctx, test.args, stdin, stdout, stderr)
			)

			if test.wantErr == nil {
				if haveErr != nil {
					t.Fatalf("error: want none, have %v", haveErr)
				}
			}

			if test.wantErr != nil {
				if !errors.Is(haveErr, test.wantErr) {
					t.Fatalf("error: want %v, have %v", test.wantErr, haveErr)
				}
			}

			{
				want := strings.TrimSpace(test.wantStdout)
				have := strings.TrimSpace(stdout.String())
				if want != have {
					t.Errorf("stdout: %s", fftest.DiffString(want, have))
				}
			}

			{
				want := strings.TrimSpace(test.wantStderr)
				have := strings.TrimSpace(stderr.String())
				if want != have {
					t.Errorf("stderr: %s", fftest.DiffString(want, have))
				}
			}
		})
	}
}

const rootUsage = `
COMMAND
  objectctl -- control objects

USAGE
  objectctl [FLAGS] <SUBCOMMAND> ...

SUBCOMMANDS
  create   create or overwrite an object
  delete   delete an object
  list     list available objects

FLAGS
      --token STRING   secret token for object API
  -v, --verbose        log verbose output
`

const listUsage = `
COMMAND
  list -- list available objects

USAGE
  objectctl list [FLAGS]

FLAGS (list)
  -a, --atime          include last access time of each object

FLAGS (objectctl)
      --token STRING   secret token for object API
  -v, --verbose        log verbose output
`

const listOutput = `
KEY       VALUE
aardvark  A nocturnal burrowing mammal with long ears, a tubular snout, and a long extensible tongue.
apple     The fruit of any of certain other species of tree of the same genus.
beach     The shore of a body of water, especially when sandy or pebbly.
carillon  A stationary set of chromatically tuned bells in a tower.
`
