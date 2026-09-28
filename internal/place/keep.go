package place

import (
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

// Meta is what a file is apart from its content and kind.
type Meta struct {
	UID, GID uint32

	// the permission bits with setuid, setgid and sticky
	Mode uint32

	Atime, Mtime unix.Timespec

	Xattrs []Xattr
}

// Xattr is an extended attribute with its value.
type Xattr struct {
	Name  string
	Value []byte
}

// Keep gives what is at a path of the image the owner, mode, extended
// attributes and times it had where it was copied from, never following a
// link at the path.
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

	var stat unix.Stat_t
	if err := unix.Fstatat(parent, name, &stat, unix.AT_SYMLINK_NOFOLLOW); err != nil {
		return &os.PathError{Op: "stat", Path: path, Err: err}
	}

	// a link has no mode of its own, and a chmod through it would reach its
	// target
	if stat.Mode&unix.S_IFMT != unix.S_IFLNK {
		if err := unix.Fchmodat(parent, name, meta.Mode, 0); err != nil {
			return &os.PathError{Op: "chmod", Path: path, Err: err}
		}
	}

	// after the owner too, whose change clears a file capability
	at := fmt.Sprintf("/proc/self/fd/%d/%s", parent, name)
	for _, xattr := range meta.Xattrs {
		if err := unix.Lsetxattr(at, xattr.Name, xattr.Value, 0); err != nil {
			return &os.PathError{Op: "setxattr " + xattr.Name, Path: path, Err: err}
		}
	}

	if err := unix.UtimesNanoAt(parent, name, []unix.Timespec{meta.Atime, meta.Mtime}, unix.AT_SYMLINK_NOFOLLOW); err != nil {
		return &os.PathError{Op: "utimes", Path: path, Err: err}
	}

	return nil
}
