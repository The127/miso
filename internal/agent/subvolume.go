package agent

import (
	"io/fs"
	"os"

	"github.com/The127/miso/internal/fstab"
)

// ownSubvolume is the directory at the top of a file system whose own fstab
// mounts it as the root, with the lines of that fstab, or empty when there
// is none.
func ownSubvolume(top string) (string, []fstab.Entry, error) {
	// a base image's links must never lead into the agent's own root
	root, err := os.OpenRoot(top)
	if err != nil {
		return "", nil, err
	}

	defer func() { _ = root.Close() }()

	children, err := fs.ReadDir(root.FS(), ".")
	if err != nil {
		return "", nil, err
	}

	for _, child := range children {
		if !child.IsDir() {
			continue
		}

		entries, _, err := fstabIn(root, child.Name())
		if err != nil {
			return "", nil, err
		}

		if subvolume, found := fstab.RootSubvolume(entries); found && subvolume == child.Name() {
			return subvolume, entries, nil
		}
	}

	return "", nil, nil
}
