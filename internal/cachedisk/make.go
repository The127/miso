package cachedisk

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// making names a disk while it is made, after the name it will have
const making = ".making-"

// Make makes a new cache disk of a size in bytes at a path, and never over
// one that is there. Only under the Lock, because it removes what a killed
// Make left, which a running one would still be making.
func Make(path string, size int64) error {
	if err := makeDisk(path, size); err != nil {
		return fmt.Errorf("make cache disk %s: %w", path, err)
	}

	return nil
}

// makeDisk is Make without naming the disk in its errors.
func makeDisk(path string, size int64) error {
	// before anything is touched, a missing mkfs leaves all as it was
	mkfs, err := findMkfs()
	if err != nil {
		return err
	}

	if err := removeLeftovers(path); err != nil {
		return err
	}

	// the disk is made under another name, so a miso killed halfway never
	// leaves a broken one at the path
	temp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+making+"*")
	if err != nil {
		return err
	}

	defer func() { _ = os.Remove(temp.Name()) }()

	if err := temp.Close(); err != nil {
		return err
	}

	if err := format(mkfs, temp.Name(), size); err != nil {
		return err
	}

	// a link, unlike a rename, never replaces what is there
	if err := os.Link(temp.Name(), path); err != nil {
		// the temporary name in the error means nothing to a user
		return errors.Unwrap(err)
	}

	return nil
}

// removeLeftovers removes the disks a killed Make left at a path.
func removeLeftovers(path string) error {
	// by name and not by pattern, a directory's name may hold a [ or a *
	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), filepath.Base(path)+making) {
			continue
		}

		if err := os.Remove(filepath.Join(filepath.Dir(path), entry.Name())); err != nil {
			return err
		}
	}

	return nil
}
