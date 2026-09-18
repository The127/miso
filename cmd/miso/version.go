package main

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"

	"github.com/urfave/cli/v3"

	"github.com/The127/miso/internal/version"
)

var versionCommand = &cli.Command{
	Name:   "version",
	Usage:  "say which miso this is",
	Action: printVersion,
}

func printVersion(_ context.Context, command *cli.Command) error {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return errors.New("this binary carries no build information")
	}

	_, err := fmt.Fprintln(command.Root().Writer, version.Of(info))

	return err
}
