package agent

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"syscall"

	"golang.org/x/sys/unix"

	"github.com/The127/miso/internal/protocol"
	"github.com/The127/miso/internal/sandbox"
)

// mountedImage is the image's layers mounted read-only at dir, and the
// directories they are mounted from, top first.
type mountedImage struct {
	dir   string
	below []string
}

func (m mountedImage) unmount() {
	_ = syscall.Unmount(m.dir, syscall.MNT_DETACH)
}

func (a *Agent) mountImage(scratch string, layers []string) (mountedImage, error) {
	dir := filepath.Join(scratch, "image")
	if err := os.Mkdir(dir, 0o700); err != nil {
		return mountedImage{}, err
	}

	below, err := a.lowers(layers)
	if err != nil {
		return mountedImage{}, err
	}

	// overlay without an upper layer wants two below
	bottom, err := empty(scratch)
	if err != nil {
		return mountedImage{}, err
	}

	below = append(below, bottom)
	if err := mountReadOnly(dir, below); err != nil {
		return mountedImage{}, err
	}

	return mountedImage{dir: dir, below: below}, nil
}

// overImage has the tools in root run a command that sees the layers below,
// top first, which it mounts at booting, as /run/miso/image.
func overImage(ctx context.Context, root, booting string, below []string, run protocol.Run, out io.Writer) (int, error) {
	seen := filepath.Join(root, "run", "miso", "image")
	if err := os.Mkdir(seen, 0o700); err != nil {
		return 0, err
	}

	if err := mountReadOnly(booting, below); err != nil {
		return 0, err
	}

	defer func() { _ = syscall.Unmount(booting, syscall.MNT_DETACH) }()

	if err := cloneInto(booting, seen); err != nil {
		return 0, err
	}

	defer func() { _ = unix.Unmount(seen, unix.MNT_DETACH) }()

	return sandbox.Run(ctx, root, run, out)
}

// bind shows a directory at a place, which it makes, and hands back what
// takes it away again.
func bind(dir, at string) (func(), error) {
	if err := os.MkdirAll(at, 0o700); err != nil {
		return nil, err
	}

	if err := unix.Mount(dir, at, "", unix.MS_BIND, ""); err != nil {
		return nil, err
	}

	return func() { _ = unix.Unmount(at, unix.MNT_DETACH) }, nil
}

// bindOutput shows the directory a layer is made in as /run/miso/out in the
// root of the tools, where they write the file of the layer.
func bindOutput(root, dir string) (func(), error) {
	return bind(dir, filepath.Join(root, "run", "miso", "out"))
}

// cloneInto shows a mount at another place. The clone keeps the attributes
// of the mount, so a device in an image opens nothing of the builder's.
func cloneInto(mount, at string) error {
	clone, err := unix.OpenTree(unix.AT_FDCWD, mount, unix.OPEN_TREE_CLONE|unix.OPEN_TREE_CLOEXEC)
	if err != nil {
		return err
	}

	err = unix.MoveMount(clone, "", unix.AT_FDCWD, at, unix.MOVE_MOUNT_F_EMPTY_PATH)
	_ = unix.Close(clone)

	return err
}
