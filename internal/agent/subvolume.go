package agent

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/The127/miso/internal/fstab"
)

// ErrSeveralRoots is a file system with more than one subvolume whose fstab
// mounts it as the root.
var ErrSeveralRoots = errors.New("several subvolumes claim the root")

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

	var (
		claims  []string
		claimed []fstab.Entry
	)

	for _, child := range children {
		if !child.IsDir() {
			continue
		}

		entries, _, err := fstabIn(root, child.Name())
		if err != nil {
			return "", nil, err
		}

		if subvolume, found := fstab.RootSubvolume(entries); found && subvolume == child.Name() {
			claims = append(claims, subvolume)
			claimed = entries
		}
	}

	switch len(claims) {
	case 0:
		return "", nil, nil
	case 1:
		return claims[0], claimed, nil
	default:
		return "", nil, fmt.Errorf("%s: %w", strings.Join(claims, ", "), ErrSeveralRoots)
	}
}
