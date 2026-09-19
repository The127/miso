package cachedisk

import (
	"errors"
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

	return "", err
}
