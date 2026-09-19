package cachedisk

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
)

// Make makes a new cache disk of a size in bytes at a path, and never over
// one that is there.
func Make(path string, size int64) error {
	// mkfs asks nobody without a terminal, it wipes the layers
	if _, err := os.Lstat(path); err == nil {
		return fmt.Errorf("make cache disk %s: %w", path, fs.ErrExist)
	}

	// a bare number would count blocks
	kibibytes := fmt.Sprintf("%dk", size>>10)

	return exec.Command("mkfs.ext4", "-q", path, kibibytes).Run() //nolint:gosec // the path is where miso keeps its own cache
}
