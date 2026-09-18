package agent

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/The127/miso/internal/disk"
	"github.com/The127/miso/internal/fstab"
)

// mountRoot mounts the root partition of the disk with a serial read-only on
// a directory.
func mountRoot(serial, target string) error {
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
	if err := syscall.Mount(device, target, kind, syscall.MS_RDONLY, ""); err != nil {
		return err
	}

	subvolume, entries, err := ownSubvolume(target)
	if err != nil {
		return err
	}

	if subvolume == "" {
		return nil
	}

	if err := syscall.Unmount(target, 0); err != nil {
		return err
	}

	if err := syscall.Mount(device, target, kind, syscall.MS_RDONLY, "subvol="+subvolume); err != nil {
		return err
	}

	image, err := os.OpenRoot(target)
	if err != nil {
		return err
	}

	defer func() { _ = image.Close() }()

	for _, submount := range fstab.Submounts(entries) {
		// the other options tune a running system and mean nothing to a copy
		var data string

		for _, option := range submount.Options {
			if strings.HasPrefix(option, "subvol=") {
				data = option
			}
		}

		// opened within the image, so a link cannot lead the mount out of it
		point, err := image.Open(strings.TrimPrefix(submount.Target, "/"))
		if err != nil {
			return err
		}

		err = syscall.Mount(device, fmt.Sprintf("/proc/self/fd/%d", point.Fd()), kind, syscall.MS_RDONLY, data)
		_ = point.Close()

		if err != nil {
			return err
		}
	}

	return nil
}
