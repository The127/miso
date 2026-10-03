package main

import (
	"context"
	"io"

	"github.com/urfave/cli/v3"

	"github.com/The127/miso/internal/build"
	"github.com/The127/miso/internal/builder"
	"github.com/The127/miso/internal/contextfiles"
	"github.com/The127/miso/internal/imagefile"
	"github.com/The127/miso/internal/plan"
	"github.com/The127/miso/internal/protocol"
	"github.com/The127/miso/internal/qemu"
)

var shellCommand = &cli.Command{
	Name:      "shell",
	Usage:     "open a shell on the layers as they are before a step of the build file",
	ArgsUsage: "[context]",
	Flags: []cli.Flag{
		fileFlag(),
		&cli.IntFlag{Name: "before", Usage: "the line of the build file whose step the shell stands before", Required: true},
	},
	Action: runShell,
}

func runShell(ctx context.Context, command *cli.Command) error {
	line := command.Int("before")
	before := func(planned plan.Plan, network protocol.Network) ([]build.Request, error) {
		return build.Before(planned, network, line)
	}

	return withBuilder(ctx, command, before, func(have prepared, boot booter) error {
		return boot(func(vm *qemu.VM, dial func() (io.ReadWriteCloser, error)) error {
			code, err := builder.Shell(ctx, vm, dial, agentName(), have.requests, contextfiles.Of(have.files), command.Root().Reader, command.Root().Writer)
			if err != nil {
				return imagefile.InFile(have.file, err)
			}

			if code != 0 {
				return cli.Exit("", code)
			}

			return nil
		})
	})
}
