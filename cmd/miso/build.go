package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v3"

	"github.com/The127/miso/internal/baseimage"
	"github.com/The127/miso/internal/build"
	"github.com/The127/miso/internal/buildcontext"
	"github.com/The127/miso/internal/builder"
	"github.com/The127/miso/internal/builderkernel"
	"github.com/The127/miso/internal/cachedisk"
	"github.com/The127/miso/internal/download"
	"github.com/The127/miso/internal/plan"
	"github.com/The127/miso/internal/protocol"
	"github.com/The127/miso/internal/qemu"
	"github.com/The127/miso/internal/vsock"
)

// cacheDiskSize is how large the cache disk may grow. The file is sparse,
// it takes only what the layers on it take.
const cacheDiskSize = 20 << 30

var buildCommand = &cli.Command{
	Name:      "build",
	Usage:     "build what the build file describes, in the builder VM",
	ArgsUsage: "[context]",
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "file", Aliases: []string{"f"}, Usage: "the build file, Imagefile in the context if not given"},
	},
	Action: runBuild,
}

func runBuild(ctx context.Context, command *cli.Command) error {
	cache, err := cacheDir()
	if err != nil {
		return err
	}

	blobs := download.Open(filepath.Join(cache, "bases"), http.DefaultClient)
	bases := baseimage.Open(filepath.Join(cache, "bases"), blobs, baseimage.Known)

	card, network, err := hostNetwork()
	if err != nil {
		return err
	}

	requests, err := planned(ctx, command, bases, network)
	if err != nil {
		return err
	}

	dir := filepath.Join(cache, "builder")
	held, err := lockedCache(dir, command.Root().ErrWriter)
	if err != nil {
		return err
	}

	defer func() { _ = held.Close() }()

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

	return inBuilder(ctx, machine, filepath.Join(dir, "builder.log"), func(vm *qemu.VM) error {
		dial := func() (io.ReadWriteCloser, error) { return vsock.Dial(vm.CID(), vsock.AgentPort) }

		return builder.Ask(ctx, vm, dial, agentName(), requests, command.Root().Writer)
	})
}

// hostNetwork is the builder's card and the network its runs get, with a
// nameserver in the families the host resolves in.
func hostNetwork() (qemu.Card, protocol.Network, error) {
	var resolving builder.Resolving

	conf, err := os.Open("/etc/resolv.conf")
	switch {
	case errors.Is(err, fs.ErrNotExist):
		// QEMU finds no resolver either
	case err != nil:
		return qemu.Card{}, protocol.Network{}, err
	default:
		defer func() { _ = conf.Close() }()

		resolving, err = builder.HostResolving(conf)
		if err != nil {
			return qemu.Card{}, protocol.Network{}, err
		}
	}

	card, network := builder.Network(resolving)

	return card, network, nil
}

// planned is what the agent is asked for the build file, with every base
// image it names fetched.
func planned(ctx context.Context, command *cli.Command, bases *baseimage.Cache, network protocol.Network) ([]protocol.Message, error) {
	dir, file := located(command)

	stages, err := parsed(file)
	if err != nil {
		return nil, err
	}

	files, err := buildcontext.Open(dir)
	if err != nil {
		return nil, err
	}

	defer func() { _ = files.Close() }()

	planning, err := plan.New(stages, agentName(), files, bases)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", file, err)
	}

	if len(planning.Downloads) > 0 {
		for _, name := range planning.Downloads {
			if _, err := bases.Fetch(ctx, name); err != nil {
				return nil, fmt.Errorf("fetch %s: %w", name, err)
			}
		}

		planning, err = plan.New(stages, agentName(), files, bases)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", file, err)
		}
	}

	return build.Requests(planning, network)
}

// lockedCache holds the cache disk in a directory for this build, and makes
// the disk if it is not there yet.
func lockedCache(dir string, said io.Writer) (io.Closer, error) {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return nil, err
	}

	held, err := cachedisk.Lock(filepath.Join(dir, "layers.lock"), func() {
		_, _ = fmt.Fprintln(said, "miso: waiting for another build to let go of the cache")
	})
	if err != nil {
		return nil, err
	}

	disk := filepath.Join(dir, "layers.img")
	if _, err := os.Stat(disk); errors.Is(err, fs.ErrNotExist) {
		err = cachedisk.Make(disk, cacheDiskSize)
		if err != nil {
			return nil, errors.Join(err, held.Close())
		}
	}

	return held, nil
}

// bootFiles writes the builder kernel and an initramfs with this very miso
// as its init into a directory of their own.
func bootFiles(ctx context.Context, blobs *download.Store) (builder.Boot, error) {
	kernel, err := builderkernel.Ready(ctx, blobs)
	if err != nil {
		return builder.Boot{}, err
	}

	self, err := os.Executable()
	if err != nil {
		return builder.Boot{}, err
	}

	init, err := os.ReadFile(self)
	if err != nil {
		return builder.Boot{}, err
	}

	dir, err := os.MkdirTemp("", "miso-boot-")
	if err != nil {
		return builder.Boot{}, err
	}

	boot, err := builder.WriteBoot(dir, kernel, init)
	if err != nil {
		return builder.Boot{}, errors.Join(err, os.RemoveAll(dir))
	}

	return boot, nil
}

// inBuilder does the work with the builder VM booted, its console in a log
// file, and stops the VM afterwards.
func inBuilder(ctx context.Context, machine qemu.Machine, log string, work func(vm *qemu.VM) error) error {
	console, err := os.Create(log)
	if err != nil {
		return err
	}

	defer func() { _ = console.Close() }()

	machine.Console = console

	running, stop := context.WithCancel(ctx)
	defer stop()

	vm, err := qemu.Driver{Binary: "qemu-system-x86_64"}.Start(running, machine)
	if err != nil {
		return err
	}

	err = work(vm)
	stop()
	<-vm.Done()

	if err != nil {
		return fmt.Errorf("%w (the builder VM's console is in %s)", err, log)
	}

	return nil
}
