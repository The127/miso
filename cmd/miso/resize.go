package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

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
	size := int64(cacheDiskSize)
	if command.IsSet("size") {
		asked, err := sizeOf("--size", command.String("size"))
		if err != nil {
			return err
		}

		size = asked
	}

	cache, err := cacheDir()
	if err != nil {
		return err
	}

	dir := filepath.Join(cache, "builder")
	disk := filepath.Join(dir, "layers.img")

	// the lock makes its file, so a missing disk is found before it
	if _, err := os.Stat(disk); errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("there is no cache disk at %s, the next build makes one of %s", disk, ibytes(cacheDiskSize))
	}

	held, err := holdCache(dir, command.Root().ErrWriter)
	if err != nil {
		return err
	}

	defer func() { _ = held.Close() }()

	// a build or another resize may have changed the disk while this one waited
	info, err := os.Stat(disk)
	if err != nil {
		return err
	}

	// no size was set: a disk that is large enough is what the user wants
	if !command.IsSet("size") && info.Size() >= size {
		_, err = fmt.Fprintf(command.Root().Writer, "the cache disk is %s already\n", ibytes(info.Size()))

		return err
	}

	if err := cachedisk.Grow(disk, size); err != nil {
		return err
	}

	_, err = fmt.Fprintf(command.Root().Writer, "the cache disk is now %s, it was %s\n", ibytes(size), ibytes(info.Size()))

	return err
}
