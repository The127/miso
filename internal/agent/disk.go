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

// repart makes the disk of the image from the definitions the image ships.
// Without --dry-run=no it writes nothing. OVMF cannot read an ESP of 4096
// byte sectors, which repart may pick for a file.
const repart = `systemd-repart \
	--dry-run=no \
	--root=/run/miso/image \
	--offline=yes \
	--sector-size=512 \
	--empty=create \
	--size=auto \
	/run/miso/out/disk.raw`

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

	// the ESP lies on the image as its efi, where the image's repart
	// definitions take it from
	esp := filepath.Join(scratch, "esp")
	fallback := filepath.Join(esp, "efi", "EFI", "BOOT")
	if err := os.MkdirAll(fallback, 0o755); err != nil { //nolint:gosec // an image's directories are open to all
		return err
	}

	loader, err := os.ReadFile(filepath.Join(image, "usr", "lib", "systemd", "boot", "efi", "systemd-bootx64.efi"))
	if err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(fallback, "BOOTX64.EFI"), loader, 0o644); err != nil { //nolint:gosec // an image's files are open to all
		return err
	}

	booting := filepath.Join(scratch, "booting")
	if err := os.Mkdir(booting, 0o700); err != nil {
		return err
	}

	if err := mountReadOnly(booting, append([]string{esp}, append(lowers, bottom)...)); err != nil {
		return err
	}

	defer func() { _ = syscall.Unmount(booting, syscall.MNT_DETACH) }()

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

		seen := filepath.Join(root, "run", "miso", "image")
		if err := os.Mkdir(seen, 0o700); err != nil {
			return false, err
		}

		// the clone keeps the attributes of the image's mount, so a device in
		// the image opens nothing of the builder's
		clone, err := unix.OpenTree(unix.AT_FDCWD, booting, unix.OPEN_TREE_CLONE|unix.OPEN_TREE_CLOEXEC)
		if err != nil {
			return false, err
		}

		err = unix.MoveMount(clone, "", unix.AT_FDCWD, seen, unix.MOVE_MOUNT_F_EMPTY_PATH)
		_ = unix.Close(clone)
		if err != nil {
			return false, err
		}

		defer func() { _ = unix.Unmount(seen, unix.MNT_DETACH) }()

		code, err = sandbox.Run(ctx, root, protocol.Run{Command: repart}, out)

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
