package main

import (
	"context"
	"errors"
	"os"

	"github.com/urfave/cli/v3"
)

var agentCommand = &cli.Command{
	Name:   "agent",
	Usage:  "answer the host, as the init of the builder VM",
	Hidden: true,
	Action: runAgent,
}

func runAgent(context.Context, *cli.Command) error {
	if os.Getpid() != 1 {
		return errors.New("miso agent runs as the init of the builder VM")
	}

	return nil
}
