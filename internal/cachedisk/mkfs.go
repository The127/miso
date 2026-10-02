package cachedisk

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
)

// searchAlso is where the tools of e2fsprogs live when the PATH leaves them
// out, as a user's PATH on Debian leaves out the sbin directories
var searchAlso = []string{"/usr/sbin", "/sbin"}

// findTool finds a tool of e2fsprogs on the PATH, or where the PATH leaves it
// out.
func findTool(name string) (string, error) {
	tool, err := exec.LookPath(name)
	if !errors.Is(err, exec.ErrNotFound) {
		return tool, err
	}

	for _, dir := range searchAlso {
		if found, errThere := exec.LookPath(filepath.Join(dir, name)); errThere == nil {
			return found, nil
		}
	}

	return "", fmt.Errorf("%w, it comes with e2fsprogs", err)
}

// format has mkfs make an ext4 of a size in bytes at a path.
func format(mkfs, path string, size int64) error {
	// a bare number would count blocks
	kibibytes := fmt.Sprintf("%dk", size>>10)

	return runTool(mkfs, "-q", path, kibibytes)
}

// runTool runs a tool of e2fsprogs on a cache disk, and adds what it said to
// its error.
func runTool(tool string, args ...string) error {
	said, err := exec.Command(tool, args...).CombinedOutput() //nolint:gosec // the path is where miso keeps its own cache
	if err != nil {
		return fmt.Errorf("%w: %s", err, bytes.TrimSpace(said))
	}

	return nil
}
