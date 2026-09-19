package basemount

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"

	"github.com/The127/miso/internal/disk"
	"github.com/The127/miso/internal/fstab"
)

// Mount mounts the root partition of the disk with a serial read-only on
// a directory.
func Mount(serial, target string) error {
	block := os.DirFS("/sys/block")

	name, err := disk.BySerial(block, serial)
	if err != nil {
		return err
	}

	f, err := os.Open(filepath.Join("/dev", name))
	if err != nil {
		return err
	}

	defer func() { _ = f.Close() }()

	root, err := disk.Root(f)
	if err != nil {
		return err
	}

	kind, err := disk.FileSystem(io.NewSectionReader(f, root.Offset, root.Size))
	if err != nil {
		return err
	}

	partition, err := disk.PartitionName(block, name, root.Number)
	if err != nil {
		return err
	}

	device := filepath.Join("/dev", partition)
	if err := mountReadOnly(device, target, kind, "", target); err != nil {
		return err
	}

	if err := assembleRoot(device, kind, target); err != nil {
		// one lazy unmount takes the submounts with it
		_ = syscall.Unmount(target, syscall.MNT_DETACH)

		return err
	}

	return nil
}

// assembleRoot turns the file system of a device mounted on a directory into
// the root that its fstab describes.
func assembleRoot(device, kind, target string) error {
	subvolume, entries, err := describedRoot(target)
	if err != nil {
		return err
	}

	if subvolume != "" {
		if err := syscall.Unmount(target, 0); err != nil {
			return err
		}

		if err := mountReadOnly(device, target, kind, "subvol="+subvolume, target); err != nil {
			return err
		}
	}

	return mountSubmounts(device, kind, target, fstab.Submounts(entries))
}

// describedRoot is the subvolume of the file system mounted on a directory
// that is the root, empty for the mount itself, and the lines of the root's
// fstab.
func describedRoot(target string) (string, []fstab.Entry, error) {
	// a base image's links must never lead into the agent's own root
	top, err := os.OpenRoot(target)
	if err != nil {
		return "", nil, err
	}

	// closed before any remount, which an open root would keep busy
	defer func() { _ = top.Close() }()

	entries, found, err := fstabIn(top, ".")
	if err != nil {
		return "", nil, err
	}

	if found {
		return "", entries, nil
	}

	return ownSubvolume(top)
}

// mountReadOnly mounts a device read-only and names in a failure what went
// where.
func mountReadOnly(device, target, kind, data, where string) error {
	if err := syscall.Mount(device, target, kind, syscall.MS_RDONLY, data); err != nil {
		return fmt.Errorf("mount %s %s on %s: %w", device, data, where, err)
	}

	return nil
}
