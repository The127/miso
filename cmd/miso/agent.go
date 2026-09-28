package main

import (
	"context"
	"errors"
	"io/fs"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/The127/miso/internal/agent"
	"github.com/The127/miso/internal/boot"
	"github.com/The127/miso/internal/guestport"
	"github.com/The127/miso/internal/protocol"
	"github.com/The127/miso/internal/qemu"
	"github.com/The127/miso/internal/vport"
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

	listener, err := listen()
	if err != nil {
		return err
	}

	return protocol.Serve(listener, agentName(), agent.Serving(protocol.CacheSerial, "/cache"))
}

// listen takes the host's connections on the virtio port a host without
// vsock gave the VM, or else over vsock.
func listen() (protocol.Listener, error) {
	device, err := guestport.Find("/sys", "/dev", qemu.AgentPort)
	if errors.Is(err, fs.ErrNotExist) {
		return vsock.Listen(vsock.AgentPort)
	}

	if err != nil {
		return nil, err
	}

	port, err := os.OpenFile(device, os.O_RDWR, 0)
	if err != nil {
		return nil, err
	}

	return vport.Listen(port)
}
