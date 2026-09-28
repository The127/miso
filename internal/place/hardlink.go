package place

import (
	"os"

	"golang.org/x/sys/unix"
)

// HardLink gives the file at an existing path of the image another name, at
// path.
func (r *Root) HardLink(path, existing string) error {
	from, fromName, err := r.parent(existing)
	if err != nil {
		return err
	}

	defer func() { _ = unix.Close(from) }()

	parent, name, err := r.parent(path)
	if err != nil {
		return err
	}

	defer func() { _ = unix.Close(parent) }()

	if err := vacate(parent, name, path); err != nil {
		return err
	}

	// never following a link at the existing name, whose target could lead
	// out of the image
	if err := unix.Linkat(from, fromName, parent, name, 0); err != nil {
		return &os.PathError{Op: "link", Path: path, Err: err}
	}

	return nil
}
