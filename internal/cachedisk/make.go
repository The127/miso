package cachedisk

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Make makes a new cache disk of a size in bytes at a path, and never over
// one that is there.
func Make(path string, size int64) error {
	// the disk is made under another name, so a miso killed halfway never
	// leaves a broken one at the path
	temp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".making-*")
	if err != nil {
		return fmt.Errorf("make cache disk %s: %w", path, err)
	}

	defer func() { _ = os.Remove(temp.Name()) }()

	if err := temp.Close(); err != nil {
		return fmt.Errorf("make cache disk %s: %w", path, err)
	}

	// a bare number would count blocks
	kibibytes := fmt.Sprintf("%dk", size>>10)

	mkfs, err := findMkfs()
	if errors.Is(err, exec.ErrNotFound) {
		return fmt.Errorf("make cache disk %s: %w, it comes with e2fsprogs", path, err)
	}

	if err != nil {
		return fmt.Errorf("make cache disk %s: %w", path, err)
	}

	said, err := exec.Command(mkfs, "-q", temp.Name(), kibibytes).CombinedOutput() //nolint:gosec // the path is where miso keeps its own cache
	if err != nil {
		return fmt.Errorf("make cache disk %s: %w: %s", path, err, bytes.TrimSpace(said))
	}

	// a link, unlike a rename, never replaces what is there
	if err := os.Link(temp.Name(), path); err != nil {
		return fmt.Errorf("make cache disk %s: %w", path, errors.Unwrap(err))
	}

	return nil
}
