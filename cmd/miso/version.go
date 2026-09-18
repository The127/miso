package main

import (
	"context"
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
	_, err := fmt.Fprintln(command.Root().Writer, built())

	return err
}

// built is the version of this binary. A binary can carry no build
// information at all, and then that is all there is to say.
func built() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}

	return version.Of(info)
}
