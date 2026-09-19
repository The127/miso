package main

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
)

func main() {
	miso := &cli.Command{
		Name:     "miso",
		Usage:    "build systemd-based operating systems from a build file",
		Commands: []*cli.Command{planCommand, versionCommand, agentCommand},
	}

	if err := miso.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, "miso:", err)
		os.Exit(1)
	}
}
