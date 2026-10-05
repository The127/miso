package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/The127/miso/internal/build"
	"github.com/The127/miso/internal/builder"
	"github.com/The127/miso/internal/contextfiles"
	"github.com/The127/miso/internal/imagefile"
	"github.com/The127/miso/internal/qemu"
)

const buildArch = "amd64"

var buildCommand = &cli.Command{
	Name:      "build",
	Usage:     "build what the build file describes, in the builder VM",
	ArgsUsage: "[context]",
	Flags: []cli.Flag{
		fileFlag(),
		outputFlag(),
		&cli.StringFlag{Name: "arch", Usage: "the architecture the image is built for"},
	},
	Action: runBuild,
}

func runBuild(ctx context.Context, command *cli.Command) error {
	arch := command.String("arch")
	if command.IsSet("arch") && arch != buildArch {
		return fmt.Errorf("--arch %q is not an architecture miso builds for, only %s", arch, buildArch)
	}

	return withBuilder(ctx, command, build.Requests, func(have prepared, boot booter) error {
		// a checked disk's console, kept for when a check fails
		console, err := os.Create(have.dir.CheckLog())
		if err != nil {
			return err
		}

		defer func() { _ = console.Close() }()

		// where disks go that are only checked, not written out
		scratch, err := os.MkdirTemp(string(have.dir), "outputs-")
		if err != nil {
			return err
		}

		defer func() { _ = os.RemoveAll(scratch) }()

		outputs, err := outputsOf(command, checker(ctx, have.blobs, string(have.dir), console), scratch)
		if err != nil {
			return err
		}

		return boot(func(vm *qemu.VM, dial func() (io.ReadWriteCloser, error)) error {
			err := builder.Ask(ctx, vm, dial, agentName(), have.requests, contextfiles.Of(have.files), outputs, command.Root().Writer)

			return withShellHint(command, have.requests, imagefile.InFile(have.file, err))
		})
	})
}
