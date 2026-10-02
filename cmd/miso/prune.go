package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/The127/miso/internal/build"
	"github.com/The127/miso/internal/builder"
	"github.com/The127/miso/internal/download"
	"github.com/The127/miso/internal/lru"
	"github.com/The127/miso/internal/protocol"
	"github.com/The127/miso/internal/qemu"
)

var pruneCommand = &cli.Command{
	Name:  "prune",
	Usage: "remove cached layers and downloads that were not used for a while",
	Flags: []cli.Flag{
		&cli.DurationFlag{Name: "older-than", Usage: "remove what no build used for longer than this, like 720h"},
		&cli.StringFlag{Name: "keep-storage", Usage: "remove the least recently used until the layers, and the downloads, each take no more than this, like 10GiB"},
	},
	Action: runPrune,
}

func runPrune(ctx context.Context, command *cli.Command) error {
	limit, err := limits(command)
	if err != nil {
		return err
	}

	cache, err := cacheDir()
	if err != nil {
		return err
	}

	blobs, _ := baseImages(cache)

	dir := filepath.Join(cache, "builder")
	held, err := lockedCache(dir, command.Root().ErrWriter)
	if err != nil {
		return err
	}

	defer func() { _ = held.Close() }()

	if err = pruneLayers(ctx, command, dir, blobs, limit); err != nil {
		return err
	}

	swept, err := blobs.Prune(lru.Policy{Now: time.Now(), OlderThan: limit.OlderThan, KeepStorage: limit.KeepStorage})
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(command.Root().Writer, "removed %d downloads, freed %d bytes\n", swept.Count, swept.Bytes)

	return err
}

// pruneLayers has the builder VM remove the layers the limit lets go, from
// the cache disk in dir, which the caller holds.
func pruneLayers(ctx context.Context, command *cli.Command, dir string, blobs *download.Store, limit protocol.Prune) error {
	boot, err := bootFiles(ctx, blobs)
	if err != nil {
		return err
	}

	defer func() { _ = os.RemoveAll(filepath.Dir(boot.Kernel)) }()

	machine, err := builder.Build{Boot: boot, Cache: filepath.Join(dir, "layers.img")}.Machine()
	if err != nil {
		return err
	}

	// a prune reaches nothing outside the VM
	machine.Card = nil

	requests := []build.Request{{Message: limit}}

	return inBuilder(ctx, machine, filepath.Join(dir, "builder.log"), command.Root().ErrWriter, func(vm *qemu.VM, dial func() (io.ReadWriteCloser, error)) error {
		return builder.Ask(ctx, vm, dial, agentName(), requests, nil, nil, command.Root().Writer)
	})
}

// limits are what the flags set, and zero for a flag that is not set. A
// limit that is not above zero is refused, because the agent reads a negative
// one as "everything" and zero as "not set".
func limits(command *cli.Command) (protocol.Prune, error) {
	if !command.IsSet("older-than") && !command.IsSet("keep-storage") {
		return protocol.Prune{}, errors.New("prune removes nothing without a limit: set --older-than or --keep-storage")
	}

	var limit protocol.Prune

	if command.IsSet("older-than") {
		age, err := ageLimit(command.Duration("older-than"))
		if err != nil {
			return protocol.Prune{}, err
		}

		limit.OlderThan = age
	}

	if command.IsSet("keep-storage") {
		keep, err := sizeOf("--keep-storage", command.String("keep-storage"))
		if err != nil {
			return protocol.Prune{}, err
		}

		limit.KeepStorage = keep
	}

	return limit, nil
}

func ageLimit(age time.Duration) (time.Duration, error) {
	if age <= 0 {
		return 0, fmt.Errorf("--older-than %s must be above zero", age)
	}

	return age, nil
}
