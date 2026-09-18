package agent

import (
	"io"
	"os"
	"path/filepath"
	"syscall"

	"github.com/The127/miso/internal/disk"
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

	return syscall.Mount(filepath.Join("/dev", partition), target, kind, syscall.MS_RDONLY, "")
}
