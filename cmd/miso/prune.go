package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/urfave/cli/v3"

	"github.com/The127/miso/internal/build"
	"github.com/The127/miso/internal/builder"
	"github.com/The127/miso/internal/protocol"
	"github.com/The127/miso/internal/qemu"
)

var pruneCommand = &cli.Command{
	Name:  "prune",
	Usage: "remove cached layers and downloads that were not used for a while",
	Flags: []cli.Flag{
		&cli.DurationFlag{Name: "older-than", Usage: "remove what no build used for longer than this, like 720h"},
		&cli.StringFlag{Name: "keep-storage", Usage: "remove the least recently used until what stays takes no more than this, like 10GiB"},
	},
	Action: runPrune,
}

func runPrune(ctx context.Context, command *cli.Command) error {
	age, keep, err := limits(command)
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

	requests := []build.Request{{Message: protocol.Prune{
		OlderThan:   age,
		KeepStorage: keep,
	}}}

	return inBuilder(ctx, machine, filepath.Join(dir, "builder.log"), command.Root().ErrWriter, func(vm *qemu.VM, dial func() (io.ReadWriteCloser, error)) error {
		return builder.Ask(ctx, vm, dial, agentName(), requests, nil, nil, command.Root().Writer)
	})
}

// limits are the age and the storage the flags set, and zero for a flag that
// is not set. A limit that is not above zero is refused, because the agent
// reads a negative one as "everything" and zero as "not set".
func limits(command *cli.Command) (time.Duration, int64, error) {
	if !command.IsSet("older-than") && !command.IsSet("keep-storage") {
		return 0, 0, errors.New("prune removes nothing without a limit: set --older-than or --keep-storage")
	}

	var age time.Duration
	var keep int64

	if command.IsSet("older-than") {
		age = command.Duration("older-than")
		if age <= 0 {
			return 0, 0, fmt.Errorf("--older-than %s must be above zero", age)
		}
	}

	if command.IsSet("keep-storage") {
		text := command.String("keep-storage")

		size, err := humanize.ParseBytes(text)
		if err != nil {
			return 0, 0, fmt.Errorf("--keep-storage %q is not a size like 10GiB: %w", text, err)
		}

		// a wrapped size would be negative
		if size > math.MaxInt64 {
			return 0, 0, fmt.Errorf("--keep-storage %q is too large", text)
		}

		if size == 0 {
			return 0, 0, fmt.Errorf("--keep-storage %q must be above zero", text)
		}

		keep = int64(size)
	}

	return age, keep, nil
}
