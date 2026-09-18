package agent

import (
	"errors"
	"io/fs"
	"os"
	"path"

	"github.com/The127/miso/internal/fstab"
)

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
