package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/dustin/go-humanize"
	"github.com/urfave/cli/v3"
)

var pruneCommand = &cli.Command{
	Name:  "prune",
	Usage: "remove cached layers and downloads that were not used for a while",
	Flags: []cli.Flag{
		&cli.DurationFlag{Name: "older-than", Usage: "remove what no build used for longer than this, like 720h"},
		&cli.StringFlag{Name: "keep-storage", Usage: "remove the least recently used until what stays takes no more than this, like 10GiB"},
	},
	Action: runPrune,
}

func runPrune(_ context.Context, command *cli.Command) error {
	if !command.IsSet("older-than") && !command.IsSet("keep-storage") {
		return errors.New("prune removes nothing without a limit: set --older-than or --keep-storage")
	}

	if command.IsSet("keep-storage") {
		if _, err := humanize.ParseBytes(command.String("keep-storage")); err != nil {
			return fmt.Errorf("--keep-storage %q is not a size like 10GiB: %w", command.String("keep-storage"), err)
		}
	}

	return nil
}
