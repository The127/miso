package cachedisk

import (
	"fmt"
	"os/exec"
)

// Make makes a new cache disk of a size in bytes at a path.
func Make(path string, size int64) error {
	// a bare number would count blocks
	kibibytes := fmt.Sprintf("%dk", size>>10)

	return exec.Command("mkfs.ext4", "-q", path, kibibytes).Run() //nolint:gosec // the path is where miso keeps its own cache
}
