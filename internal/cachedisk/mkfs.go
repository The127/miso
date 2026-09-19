package cachedisk

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
)

// searchAlso is where mkfs lives when the PATH leaves it out, as a user's
// PATH on Debian leaves out the sbin directories
var searchAlso = []string{"/usr/sbin", "/sbin"}

// findMkfs finds mkfs.ext4 on the PATH, or where the PATH leaves it out.
func findMkfs() (string, error) {
	mkfs, err := exec.LookPath("mkfs.ext4")
	if !errors.Is(err, exec.ErrNotFound) {
		return mkfs, err
	}

	for _, dir := range searchAlso {
		if found, errThere := exec.LookPath(filepath.Join(dir, "mkfs.ext4")); errThere == nil {
			return found, nil
		}
	}

	return "", fmt.Errorf("%w, it comes with e2fsprogs", err)
}

// format has mkfs make an ext4 of a size in bytes at a path.
func format(mkfs, path string, size int64) error {
	// a bare number would count blocks
	kibibytes := fmt.Sprintf("%dk", size>>10)

	said, err := exec.Command(mkfs, "-q", path, kibibytes).CombinedOutput() //nolint:gosec // the path is where miso keeps its own cache
	if err != nil {
		return fmt.Errorf("%w: %s", err, bytes.TrimSpace(said))
	}

	return nil
}
