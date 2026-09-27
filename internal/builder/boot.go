package builder

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/The127/miso/internal/builderkernel"
	"github.com/The127/miso/internal/initramfs"
)

// Boot is where the files the builder VM boots from are.
type Boot struct {
	Kernel    string
	Initramfs string
}

// WriteBoot writes the files the builder VM boots from into a directory:
// the kernel, and an initramfs with init as its init and the kernel's
// modules.
func WriteBoot(dir string, kernel builderkernel.Kernel, init []byte) (Boot, error) {
	var initrd bytes.Buffer
	if err := initramfs.Write(&initrd, init, kernel.Modules); err != nil {
		return Boot{}, err
	}

	image, err := written(dir, "vmlinuz", kernel.Image)
	if err != nil {
		return Boot{}, err
	}

	initramfsPath, err := written(dir, "initramfs", initrd.Bytes())
	if err != nil {
		return Boot{}, err
	}

	return Boot{Kernel: image, Initramfs: initramfsPath}, nil
}

// written is the path of a file with a name in a directory, once it holds
// the content given.
func written(dir, name string, content []byte) (string, error) {
	path := filepath.Join(dir, name)

	return path, os.WriteFile(path, content, 0o600)
}
