package agent

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
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

// ukify builds the UKI of the image's kernel into the ESP, with the
// image's kernel command line when it has one.
const ukify = `set --
if [ -e /run/miso/image/etc/kernel/cmdline ]; then
	set -- --cmdline=@/run/miso/image/etc/kernel/cmdline
fi
ukify build \
	--linux="/run/miso/image/$MISO_LINUX" \
	--initrd="/run/miso/image/$MISO_INITRD" \
	--stub=/run/miso/image/` + stub + ` \
	--os-release=@/run/miso/image/etc/os-release \
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

	found, err := kernel.Find(os.DirFS(image), "")
	if err != nil {
		return err
	}

	// the ESP lies on the image as its efi, where the image's repart
	// definitions take it from
	esp := filepath.Join(scratch, "esp")
	fallback := filepath.Join(esp, "efi", "EFI", "BOOT")
	if err := os.MkdirAll(fallback, 0o755); err != nil { //nolint:gosec // an image's directories are open to all
		return err
	}

	if err := os.Mkdir(filepath.Join(esp, "efi", "EFI", "Linux"), 0o755); err != nil { //nolint:gosec // an image's directories are open to all
		return err
	}

	// the image's root shows the mode and owner of the top layer, which the
	// ESP is
	root, err := os.Stat(image)
	if err != nil {
		return err
	}

	if err := os.Chmod(esp, root.Mode().Perm()); err != nil {
		return err
	}

	owner, _ := root.Sys().(*syscall.Stat_t)
	if err := os.Lchown(esp, int(owner.Uid), int(owner.Gid)); err != nil {
		return err
	}

	loader, err := fs.ReadFile(place.Open(image).FS(), systemdBoot)
	if errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("the image has no systemd-boot at /%s", systemdBoot)
	}

	if err != nil {
		return err
	}

	_, err = os.Stat(filepath.Join(image, stub))
	if errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("the image has no UKI stub at /%s", stub)
	}

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

		if err := cloneInto(image, seen); err != nil {
			return false, err
		}

		building := filepath.Join(root, "run", "miso", "esp")
		if err := os.Mkdir(building, 0o700); err != nil {
			return false, err
		}

		if err := unix.Mount(filepath.Join(esp, "efi"), building, "", unix.MS_BIND, ""); err != nil {
			_ = unix.Unmount(seen, unix.MNT_DETACH)

			return false, err
		}

		// names from the image reach the shell only as values, never as its
		// words
		uki := protocol.Run{Command: ukify, Env: []string{
			"MISO_LINUX=" + found.Linux,
			"MISO_INITRD=" + found.Initrd,
			"MISO_VERSION=" + found.Version,
		}}

		var err error
		code, err = sandbox.Run(ctx, root, uki, out)
		_ = unix.Unmount(building, unix.MNT_DETACH)
		_ = unix.Unmount(seen, unix.MNT_DETACH)
		if err != nil || code != 0 {
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

// cloneInto shows a mount at another place. The clone keeps the attributes
// of the mount, so a device in an image opens nothing of the builder's.
func cloneInto(mount, place string) error {
	clone, err := unix.OpenTree(unix.AT_FDCWD, mount, unix.OPEN_TREE_CLONE|unix.OPEN_TREE_CLOEXEC)
	if err != nil {
		return err
	}

	err = unix.MoveMount(clone, "", unix.AT_FDCWD, place, unix.MOVE_MOUNT_F_EMPTY_PATH)
	_ = unix.Close(clone)

	return err
}
