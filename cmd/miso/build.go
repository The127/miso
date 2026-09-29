package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v3"

	"github.com/The127/miso/internal/baseimage"
	"github.com/The127/miso/internal/build"
	"github.com/The127/miso/internal/buildcontext"
	"github.com/The127/miso/internal/builder"
	"github.com/The127/miso/internal/contextfiles"
	"github.com/The127/miso/internal/imagefile"
	"github.com/The127/miso/internal/protocol"
	"github.com/The127/miso/internal/qemu"
)

var buildCommand = &cli.Command{
	Name:      "build",
	Usage:     "build what the build file describes, in the builder VM",
	ArgsUsage: "[context]",
	Flags: []cli.Flag{
		fileFlag(),
		outputFlag(),
	},
	Action: runBuild,
}

func runBuild(ctx context.Context, command *cli.Command) error {
	cache, err := cacheDir()
	if err != nil {
		return err
	}

	blobs, bases := baseImages(cache)

	card, network, err := hostNetwork()
	if err != nil {
		return err
	}

	// one open context for plan and build, so the build reads the directory
	// the plan hashed even if its path changes in between
	contextDir, file := located(command)
	files, err := buildcontext.Open(contextDir)
	if err != nil {
		return err
	}

	defer func() { _ = files.Close() }()

	requests, err := planned(ctx, command, files, bases, network)
	if err != nil {
		return err
	}

	dir := filepath.Join(cache, "builder")
	held, err := lockedCache(dir, command.Root().ErrWriter)
	if err != nil {
		return err
	}

	defer func() { _ = held.Close() }()

	// a checked disk's console, kept for when a check fails
	console, err := os.Create(filepath.Join(dir, "check.log"))
	if err != nil {
		return err
	}

	defer func() { _ = console.Close() }()

	outputs, err := outputsOf(command, checker(ctx, blobs, dir, console))
	if err != nil {
		return err
	}

	boot, err := bootFiles(ctx, blobs)
	if err != nil {
		return err
	}

	defer func() { _ = os.RemoveAll(filepath.Dir(boot.Kernel)) }()

	machine, err := builder.Build{
		Boot:     boot,
		Cache:    filepath.Join(dir, "layers.img"),
		Requests: requests,
		Blob:     blobs.Path,
		Format:   bases.Format,
		Card:     card,
	}.Machine()
	if err != nil {
		return err
	}

	return inBuilder(ctx, machine, filepath.Join(dir, "builder.log"), command.Root().ErrWriter, func(vm *qemu.VM, dial func() (io.ReadWriteCloser, error)) error {
		return imagefile.InFile(file, builder.Ask(ctx, vm, dial, agentName(), requests, contextfiles.Of(files), outputs, command.Root().Writer))
	})
}

// planned is what the agent is asked for the build file, with every base
// image it names fetched.
func planned(ctx context.Context, command *cli.Command, files *buildcontext.Dir, bases *baseimage.Cache, network protocol.Network) ([]build.Request, error) {
	planning, err := planOf(command, files, bases)
	if err != nil {
		return nil, err
	}

	if len(planning.Downloads) > 0 {
		for _, name := range planning.Downloads {
			if _, err := bases.Fetch(ctx, name); err != nil {
				return nil, fmt.Errorf("fetch %s: %w", name, err)
			}
		}

		planning, err = planOf(command, files, bases)
		if err != nil {
			return nil, err
		}
	}

	return build.Requests(planning, network)
}
