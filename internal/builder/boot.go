package builder

import (
	"os"
	"path/filepath"

	"github.com/The127/miso/internal/builderkernel"
)

// Boot is where the files the builder VM boots from are.
type Boot struct {
	Kernel string
}

// WriteBoot writes the files the builder VM boots from into a directory.
func WriteBoot(dir string, kernel builderkernel.Kernel, _ []byte) (Boot, error) {
	image := filepath.Join(dir, "vmlinuz")
	if err := os.WriteFile(image, kernel.Image, 0o600); err != nil {
		return Boot{}, err
	}

	return Boot{Kernel: image}, nil
}
