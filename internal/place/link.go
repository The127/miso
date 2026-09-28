package place

import (
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

// Link puts a link with a target, as written, at a path of the image.
func (r *Root) Link(path string, target string) error {
	parent, err := r.parent(path)
	if err != nil {
		return err
	}

	defer func() { _ = unix.Close(parent) }()

	name := filepath.Base(path)
	if err := vacate(parent, name, path); err != nil {
		return err
	}

	if err := unix.Symlinkat(target, parent, name); err != nil {
		return &os.PathError{Op: "symlink", Path: path, Err: err}
	}

	return nil
}
