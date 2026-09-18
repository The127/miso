package main

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/urfave/cli/v3"

	"github.com/The127/miso/internal/version"
)

func main() {
	miso := &cli.Command{
		Name:  "miso",
		Usage: "build systemd-based operating systems from a build file",
		Commands: []*cli.Command{
			{
				Name:   "version",
				Usage:  "say which miso this is",
				Action: printVersion,
			},
		},
	}

	if err := miso.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, "miso:", err)
		os.Exit(1)
	}
}

func printVersion(_ context.Context, command *cli.Command) error {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return fmt.Errorf("this binary carries no build information")
	}

	_, err := fmt.Fprintln(command.Root().Writer, version.Of(info))

	return err
}
