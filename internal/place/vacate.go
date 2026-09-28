package place

import (
	"errors"
	"os"

	"golang.org/x/sys/unix"
)

// vacate removes what is at a name, so that what is placed there replaces it
// and takes neither a link's target nor an old file's mode or other names.
// A directory is never removed.
func vacate(parent int, name string, path string) error {
	if err := unix.Unlinkat(parent, name, 0); err != nil && !errors.Is(err, unix.ENOENT) {
		return &os.PathError{Op: "remove", Path: path, Err: err}
	}

	return nil
}
