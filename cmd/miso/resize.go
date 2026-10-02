package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"

	"github.com/The127/miso/internal/cachedisk"
)

var resizeCommand = &cli.Command{
	Name:  "resize",
	Usage: "make the cache disk larger and keep the layers on it",
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "size", Usage: "the size the cache disk gets, like 80GiB, " + ibytes(cacheDiskSize) + " when not set"},
	},
	Action: runResize,
}

func runResize(_ context.Context, command *cli.Command) error {
	size, asked, err := wantedSize(command)
	if err != nil {
		return err
	}

	cache, err := cacheDir()
	if err != nil {
		return err
	}

	dir := builderDir(cache)

	held, info, err := lockedDisk(dir, command.Root().ErrWriter)
	if err != nil {
		return err
	}

	defer func() { _ = held.Close() }()

	// no size was set: a disk that is large enough is what the user wants
	if !asked && info.Size() >= size {
		_, err = fmt.Fprintf(command.Root().Writer, "the cache disk is %s already\n", ibytes(info.Size()))

		return err
	}

	if err := cachedisk.Grow(dir.Disk(), size); err != nil {
		return err
	}

	_, err = fmt.Fprintf(command.Root().Writer, "the cache disk is now %s, it was %s\n", ibytes(size), ibytes(info.Size()))

	return err
}

// wantedSize says with asked that the user chose the size, because then a
// disk that is larger is not left alone.
func wantedSize(command *cli.Command) (size int64, asked bool, err error) {
	if !command.IsSet("size") {
		return cacheDiskSize, false, nil
	}

	size, err = sizeOf("--size", command.String("size"))

	return size, true, err
}
