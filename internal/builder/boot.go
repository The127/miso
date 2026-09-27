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
	image := filepath.Join(dir, "vmlinuz")
	if err := os.WriteFile(image, kernel.Image, 0o600); err != nil {
		return Boot{}, err
	}

	var initrd bytes.Buffer
	if err := initramfs.Write(&initrd, init, kernel.Modules); err != nil {
		return Boot{}, err
	}

	initramfsPath := filepath.Join(dir, "initramfs")
	if err := os.WriteFile(initramfsPath, initrd.Bytes(), 0o600); err != nil {
		return Boot{}, err
	}

	return Boot{Kernel: image, Initramfs: initramfsPath}, nil
}
