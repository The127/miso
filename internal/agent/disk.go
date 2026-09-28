package agent

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"

	"golang.org/x/sys/unix"

	"github.com/The127/miso/internal/kernel"
	"github.com/The127/miso/internal/protocol"
	"github.com/The127/miso/internal/sandbox"
)

// Disk makes a bootable disk image of the image's layers with the tools of
// another stage.
func (a *Agent) Disk(ctx context.Context, request protocol.Disk, out io.Writer) error {
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

	work, err := a.layers.Begin(request.Key)
	if err != nil {
		return err
	}

	// the tools run as a RUN does, but nothing they write is kept
	upper := filepath.Join(scratch, "tools")
	if err := os.Mkdir(upper, 0o700); err != nil {
		return err
	}

	code := 0
	err = a.overlaid(request.Tools, sandbox.Floor, nil, upper, func(root string) (bool, error) {
		output := filepath.Join(root, "run", "miso", "out")
		if err := os.MkdirAll(output, 0o700); err != nil {
			return false, err
		}

		if err := unix.Mount(work.Dir(), output, "", unix.MS_BIND, ""); err != nil {
			return false, err
		}

		defer func() { _ = unix.Unmount(output, unix.MNT_DETACH) }()

		var err error
		code, err = sandbox.Run(ctx, root, protocol.Run{Command: "systemd-repart /run/miso/out/disk.raw"}, out)

		return false, err
	})
	if err == nil && code != 0 {
		err = fmt.Errorf("the tools failed making the disk: exit code %d", code)
	}

	if err != nil {
		_ = work.Discard()

		return err
	}

	return work.Finish()
}
