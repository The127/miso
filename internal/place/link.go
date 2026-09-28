package place

import (
	"os"

	"golang.org/x/sys/unix"
)

// Link puts a link with a target, as written, at a path of the image.
func (r *Root) Link(path string, target string) error {
	parent, name, err := r.parent(path)
	if err != nil {
		return err
	}

	if name == "" {
		return &os.PathError{Op: "symlink", Path: path, Err: unix.EISDIR}
	}

	defer func() { _ = unix.Close(parent) }()

	if err := vacate(parent, name, path); err != nil {
		return err
	}

	if err := unix.Symlinkat(target, parent, name); err != nil {
		return &os.PathError{Op: "symlink", Path: path, Err: err}
	}

	// the link itself, a chown that follows it would reach its target
	if err := unix.Fchownat(parent, name, 0, 0, unix.AT_SYMLINK_NOFOLLOW); err != nil {
		return &os.PathError{Op: "chown", Path: path, Err: err}
	}

	if err := unix.UtimesNanoAt(parent, name, make([]unix.Timespec, 2), unix.AT_SYMLINK_NOFOLLOW); err != nil {
		return &os.PathError{Op: "utimes", Path: path, Err: err}
	}

	return nil
}
