package main

import (
	"bufio"
	"flag"
	"log"
	"os"
	"strings"

	"github.com/mitchellh/go-wordwrap"
	"github.com/peterbourgon/ff"
	"github.com/peterbourgon/ff/ffcli"
	"golang.org/x/xerrors"
)

func main() {
	log.SetFlags(0)

	var (
		globalfs        = flag.NewFlagSet("objectctl", flag.ExitOnError)
		globalToken     = globalfs.String("token", "", "access token (required; or via OBJECTCTL_TOKEN)")
		createfs        = flag.NewFlagSet("create", flag.ExitOnError)
		createPrecision = createfs.Float64("precision", 0.50, "precision of created object")
	)

	cache := &cache{
		token: "SECRET",
	}

	create := &ffcli.Command{
		Name:      "create",
		Usage:     "objectctl create [flags] <object ID>",
		ShortHelp: "create an object",
		FlagSet:   createfs,
		Exec: func(args []string) error {
			if len(args) != 1 {
				return xerrors.New("create requires 1 argument")
			}
			return cache.create(*globalToken, args[0], *createPrecision)
		},
	}

	delete := &ffcli.Command{
		Name:      "delete",
		Usage:     "objectctl delete [object ID]",
		ShortHelp: "delete an object, or all objects",
		LongHelp: collapse(`
			Delete with a single argument deletes the specified object ID.
			Delete without any arguments recursively deletes all objects.
			Obviously, be careful.
		`, 60),
		Exec: func(args []string) error {
			if len(args) <= 0 {
				return cache.deleteAll(*globalToken)
			}
			return cache.delete(*globalToken, args[0])
		},
	}

	root := &ffcli.Command{
		Usage:   "objectctl [global flags] <subcommand> [flags] [args...]",
		FlagSet: globalfs,
		Options: []ff.Option{ff.WithEnvVarPrefix("OBJECTCTL")},
		LongHelp: collapse(`
			Manipulate objects in some kind of theoretical object store.
			There are subcommands to create objects, delete specific objects,
			and delete all objects. More sophisticated help or usage text
			could go here.
		`, 78),
		Subcommands: []*ffcli.Command{create, delete},
		Exec:        func([]string) error { return xerrors.New("specify a subcommand") },
	}

	if err := root.Run(os.Args[1:]); err != nil {
		log.Fatalf("error: %v", err)
	}
}

//
//
//

var (
	errInvalidToken = xerrors.New("invalid token")
	errNoObject     = xerrors.New("no object ID provided")
)

type cache struct {
	token string
}

func (c *cache) create(token string, id string, precision float64) error {
	if id == "" {
		return errNoObject
	}

	if token != c.token {
		return errInvalidToken
	}

	log.Printf("creating object %q with precision %f", id, precision)
	return nil
}

func (c *cache) delete(token string, id string) error {
	if id == "" {
		return errNoObject
	}

	if token != c.token {
		return errInvalidToken
	}

	log.Printf("deleting object %q", id)
	return nil
}

func (c *cache) deleteAll(token string) error {
	if token == c.token {
		return errInvalidToken
	}

	log.Printf("deleting ALL objects")
	return nil
}

//
//
//

func collapse(body string, width uint) string {
	var b strings.Builder
	s := bufio.NewScanner(strings.NewReader(body))
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" {
			continue
		}
		b.WriteString(line + " ")
	}
	return wordwrap.WrapString(b.String(), width)
}
