package agent

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"syscall"

	"github.com/The127/miso/internal/kernel"
	"github.com/The127/miso/internal/protocol"
)

// Disk makes a bootable disk image of the image's layers with the tools of
// another stage.
func (a *Agent) Disk(_ context.Context, request protocol.Disk, _ io.Writer) error {
	scratch, err := a.layers.Scratch()
	if err != nil {
		return err
	}

	defer func() { _ = os.RemoveAll(scratch) }()

	image := filepath.Join(scratch, "image")
	if err := os.Mkdir(image, 0o700); err != nil {
		return err
	}

	lowers, err := a.lowers(request.Layers)
	if err != nil {
		return err
	}

	// overlay without an upper layer wants two below
	bottom, err := empty(scratch)
	if err != nil {
		return err
	}

	if err := mountReadOnly(image, append(lowers, bottom)); err != nil {
		return err
	}

	// removing scratch must never reach into the image, so it goes first
	defer func() { _ = syscall.Unmount(image, syscall.MNT_DETACH) }()

	if _, err := kernel.Find(os.DirFS(image), ""); err != nil {
		return err
	}

	return errors.New("the agent makes no disks yet")
}
