package main

import (
	"context"
	"errors"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/The127/miso/internal/agent"
	"github.com/The127/miso/internal/boot"
	"github.com/The127/miso/internal/protocol"
	"github.com/The127/miso/internal/vsock"
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

	if err := boot.Boot(); err != nil {
		return err
	}

	listener, err := vsock.Listen(vsock.AgentPort)
	if err != nil {
		return err
	}

	return protocol.Serve(listener, "miso "+built(), agent.Serving(protocol.CacheSerial, "/cache"))
}
