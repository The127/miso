package main

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/The127/miso/internal/sandbox"
	"github.com/The127/miso/internal/vsockns"
)

func main() {
	// the agent starts itself again as the helper of each run, which never
	// returns
	sandbox.Helper()
	// miso starts itself again as the holder of its vsock namespace
	vsockns.Helper()

	miso := &cli.Command{
		Name:     "miso",
		Usage:    "build systemd-based operating systems from a build file",
		Commands: []*cli.Command{planCommand, buildCommand, versionCommand, agentCommand},
	}

	if err := miso.Run(context.Background(), os.Args); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "miso:", err)
		os.Exit(1)
	}
}
