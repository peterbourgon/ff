package deletecmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/examples/objectctl/pkg/rootcmd"
)

type DeleteConfig struct {
	*rootcmd.RootConfig
	Force   bool
	Flags   *ff.CoreFlags
	Command *ff.Command
}

func New(parent *rootcmd.RootConfig) *DeleteConfig {
	var cfg DeleteConfig
	cfg.RootConfig = parent
	cfg.Flags = ff.NewFlags("delete").SetParent(parent.Flags)
	cfg.Flags.BoolVar(&cfg.Force, 0, "force", false, "force delete")
	cfg.Command = &ff.Command{
		Name:      "delete",
		Usage:     "objectctl delete <KEY>",
		ShortHelp: "delete an object",
		Flags:     cfg.Flags,
		Exec:      cfg.Exec,
	}
	cfg.RootConfig.Command.Subcommands = append(cfg.RootConfig.Command.Subcommands, cfg.Command)
	return &cfg
}

func (cfg *DeleteConfig) Exec(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return errors.New("delete requires at least 1 arg")
	}

	key := args[0]
	existed, err := cfg.Client.Delete(ctx, key, cfg.Force)
	if err != nil {
		return err
	}

	if cfg.Verbose {
		fmt.Fprintf(cfg.Stderr, "delete %q OK (existed %v)\n", key, existed)
	}

	return nil
}
