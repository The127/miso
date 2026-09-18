package agent

import (
	"errors"
	"io/fs"
	"os"
	"path"

	"github.com/The127/miso/internal/fstab"
)

// rootFstab is the fstab of the root mounted on a directory, if it has one.
func rootFstab(top string) ([]fstab.Entry, bool, error) {
	// a base image's links must never lead into the agent's own root
	root, err := os.OpenRoot(top)
	if err != nil {
		return nil, false, err
	}

	defer func() { _ = root.Close() }()

	return fstabIn(root, ".")
}

// fstabIn is the fstab of the tree in a directory of an image, if it has one.
func fstabIn(root *os.Root, dir string) ([]fstab.Entry, bool, error) {
	f, err := root.Open(path.Join(dir, "etc/fstab"))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, false, nil
	}

	if err != nil {
		return nil, false, err
	}

	defer func() { _ = f.Close() }()

	entries, err := fstab.Parse(f)
	if err != nil {
		return nil, false, err
	}

	return entries, true, nil
}
