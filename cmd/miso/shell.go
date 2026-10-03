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
	Usage:     "open a shell on the layers as they are before a step of the build file, or after the last",
	ArgsUsage: "[context]",
	Flags: []cli.Flag{
		fileFlag(),
		&cli.IntFlag{Name: "before", Usage: "the line of the build file whose step the shell stands before, the end of the build if not given"},
	},
	Action: runShell,
}

func runShell(ctx context.Context, command *cli.Command) error {
	shellOn := func(planned plan.Plan, network protocol.Network) ([]build.Request, error) {
		if !command.IsSet("before") {
			return build.After(planned, network)
		}

		return build.Before(planned, network, command.Int("before"))
	}

	return withBuilder(ctx, command, shellOn, func(have prepared, boot booter) error {
		return boot(func(vm *qemu.VM, dial func() (io.ReadWriteCloser, error)) error {
			terminal, like, stop, err := onUserTerminal(command.Root().Reader)
			if err != nil {
				return err
			}

			defer stop()

			code, err := builder.Shell(ctx, vm, dial, agentName(), told(have.requests, like), contextfiles.Of(have.files), terminal, command.Root().Writer)
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
