package agent

import (
	"errors"
	"io/fs"
	"os"
	"path"

	"github.com/The127/miso/internal/fstab"
)

// ownSubvolume is the directory at the top of a file system whose own fstab
// mounts it as the root, or empty when there is none.
func ownSubvolume(top string) (string, error) {
	// a base image's links must never lead into the agent's own root
	root, err := os.OpenRoot(top)
	if err != nil {
		return "", err
	}

	defer func() { _ = root.Close() }()

	children, err := fs.ReadDir(root.FS(), ".")
	if err != nil {
		return "", err
	}

	for _, child := range children {
		if !child.IsDir() {
			continue
		}

		f, err := root.Open(path.Join(child.Name(), "etc/fstab"))
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}

		if err != nil {
			return "", err
		}

		entries, err := fstab.Parse(f)
		_ = f.Close()

		if err != nil {
			return "", err
		}

		if subvolume, found := fstab.RootSubvolume(entries); found && subvolume == child.Name() {
			return subvolume, nil
		}
	}

	return "", nil
}
