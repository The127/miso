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
	"github.com/The127/miso/internal/cachedisk"
	"github.com/The127/miso/internal/download"
	"github.com/The127/miso/internal/plan"
	"github.com/The127/miso/internal/protocol"
	"github.com/The127/miso/internal/qemu"
)

// prepared is what a command that boots the builder has once the build file
// is planned and the cache is its own, before the VM boots.
type prepared struct {
	files    *buildcontext.Dir
	file     string
	blobs    *download.Store
	dir      cachedisk.Dir
	requests []build.Request
}

// inVM is work that runs against the builder VM, reaching its agent by dial.
type inVM func(vm *qemu.VM, dial func() (io.ReadWriteCloser, error)) error

// booter boots the builder VM, runs work in it and stops it again.
type booter func(work inVM) error

// requestsOf are the requests a command asks the builder for, from the plan
// of the build file.
type requestsOf func(planned plan.Plan, network protocol.Network) ([]build.Request, error)

// withBuilder plans the build file, takes the cache and hands use what it
// has before the VM boots, and a boot that runs work in the builder VM. Use
// may refuse before anything boots.
func withBuilder(ctx context.Context, command *cli.Command, requests requestsOf, use func(have prepared, boot booter) error) error {
	cache, err := cacheDir()
	if err != nil {
		return err
	}

	blobs, bases, err := baseImages(cache)
	if err != nil {
		return err
	}

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

	asked, err := planned(ctx, command, files, bases, network, requests)
	if err != nil {
		return err
	}

	dir := builderDir(cache)
	held, err := lockedCache(dir, command.Root().ErrWriter)
	if err != nil {
		return err
	}

	defer func() { _ = held.Close() }()

	boot := func(work inVM) error {
		bootable, err := bootFiles(ctx, blobs)
		if err != nil {
			return err
		}

		defer func() { _ = os.RemoveAll(filepath.Dir(bootable.Kernel)) }()

		machine, err := builder.Build{
			Boot:     bootable,
			Cache:    dir.Disk(),
			Requests: asked,
			Blob:     blobs.Path,
			Format:   bases.Format,
			Card:     card,
			Arch:     builderArch,
		}.Machine()
		if err != nil {
			return err
		}

		return inBuilder(ctx, machine, dir.BuilderLog(), command.Root().ErrWriter, work)
	}

	return use(prepared{files: files, file: file, blobs: blobs, dir: dir, requests: asked}, boot)
}

// planned is what the agent is asked for the build file, as the requests
// func makes them, with every base image it names fetched.
func planned(ctx context.Context, command *cli.Command, files *buildcontext.Dir, bases *baseimage.Cache, network protocol.Network, requests requestsOf) ([]build.Request, error) {
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

	return requests(planning, network)
}
