package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/The127/miso/internal/disk"
)

// MountCache mounts the cache disk with a serial on a directory.
func MountCache(serial, dir string) error {
	name, err := disk.BySerial(os.DirFS("/sys/block"), serial)
	if err != nil {
		return err
	}

	// a run reads the layers below it, which must not write to the disk
	if err := syscall.Mount(filepath.Join("/dev", name), dir, "ext4", syscall.MS_NOATIME, ""); err != nil {
		return fmt.Errorf("mount cache disk %s: %w", serial, err)
	}

	return nil
}
