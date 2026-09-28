package place

import (
	"os"

	"golang.org/x/sys/unix"
)

// Meta is what a file is apart from its content and kind.
type Meta struct {
	UID, GID uint32

	// the permission bits with setuid, setgid and sticky
	Mode uint32

	Atime, Mtime unix.Timespec
}

// Keep gives what is at a path of the image the owner, mode and times it
// had where it was copied from, never following a link at the path.
func (r *Root) Keep(path string, meta Meta) error {
	parent, name, err := r.parent(path)
	if err != nil {
		return err
	}

	defer func() { _ = unix.Close(parent) }()

	// first, because a new owner clears setuid and setgid
	if err := unix.Fchownat(parent, name, int(meta.UID), int(meta.GID), unix.AT_SYMLINK_NOFOLLOW); err != nil {
		return &os.PathError{Op: "chown", Path: path, Err: err}
	}

	if err := unix.Fchmodat(parent, name, meta.Mode, 0); err != nil {
		return &os.PathError{Op: "chmod", Path: path, Err: err}
	}

	if err := unix.UtimesNanoAt(parent, name, []unix.Timespec{meta.Atime, meta.Mtime}, unix.AT_SYMLINK_NOFOLLOW); err != nil {
		return &os.PathError{Op: "utimes", Path: path, Err: err}
	}

	return nil
}
