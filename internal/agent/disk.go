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
	"github.com/The127/miso/internal/place"
	"github.com/The127/miso/internal/protocol"
	"github.com/The127/miso/internal/sandbox"
)

// systemdBoot is where an image brings systemd-boot.
const systemdBoot = "usr/lib/systemd/boot/efi/systemd-bootx64.efi"

// stub is where an image brings the stub a UKI is built on.
const stub = "usr/lib/systemd/boot/efi/linuxx64.efi.stub"

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

// ukify builds the UKI of the image's kernel into the ESP from copies of
// the image's parts, with its kernel command line when it has one.
const ukify = `set --
if [ -e /run/miso/parts/cmdline ]; then
	set -- --cmdline=@/run/miso/parts/cmdline
fi
ukify build \
	--linux=/run/miso/parts/linux \
	--initrd=/run/miso/parts/initrd \
	--stub=/run/miso/parts/stub \
	--os-release=@/run/miso/parts/os-release \
	--uname="$MISO_VERSION" \
	--output="/run/miso/esp/EFI/Linux/$MISO_VERSION.efi" \
	"$@"`

// Disk makes a bootable disk image of the image's layers with the tools of
// another stage. A key whose layer is there already has its disk.
func (a *Agent) Disk(ctx context.Context, request protocol.Disk, out io.Writer) error {
	there, err := a.layers.Has(request.Key)
	if err != nil || there {
		return err
	}

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

	// links in the image mean places in the image, never in the builder VM
	imageFS := place.Open(image).FS()

	found, err := kernel.Find(imageFS, "")
	if err != nil {
		return err
	}

	esp := filepath.Join(scratch, "esp")
	if err := makeESP(imageFS, image, esp); err != nil {
		return err
	}

	parts := filepath.Join(scratch, "parts")
	if err := os.Mkdir(parts, 0o700); err != nil {
		return err
	}

	if err := copyParts(imageFS, found, parts); err != nil {
		return err
	}

	booting := filepath.Join(scratch, "booting")
	if err := os.Mkdir(booting, 0o700); err != nil {
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
		unbindOutput, err := bind(work.Dir(), filepath.Join(root, "run", "miso", "out"))
		if err != nil {
			return false, err
		}

		defer unbindOutput()

		unbindParts, err := bind(parts, filepath.Join(root, "run", "miso", "parts"))
		if err != nil {
			return false, err
		}

		unbindESP, err := bind(filepath.Join(esp, "efi"), filepath.Join(root, "run", "miso", "esp"))
		if err != nil {
			unbindParts()

			return false, err
		}

		// a version from the image reaches the shell only as a value, never
		// as its words
		uki := protocol.Run{Command: ukify, Env: []string{"MISO_VERSION=" + found.Version}}

		code, err = sandbox.Run(ctx, root, uki, out)
		unbindESP()
		unbindParts()
		if err != nil || code != 0 {
			return false, err
		}

		seen := filepath.Join(root, "run", "miso", "image")
		if err := os.Mkdir(seen, 0o700); err != nil {
			return false, err
		}

		// the ESP is whole now, and no lower layer may change under a
		// mounted overlay
		if err := mountReadOnly(booting, append([]string{esp}, append(lowers, bottom)...)); err != nil {
			return false, err
		}

		defer func() { _ = syscall.Unmount(booting, syscall.MNT_DETACH) }()

		if err := cloneInto(booting, seen); err != nil {
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
