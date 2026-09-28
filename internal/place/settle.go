package place

import (
	"os"

	"golang.org/x/sys/unix"
)

// settle gives what was placed the owner, mode and time that a digest
// implies, whoever and whenever placed it.
func settle(fd int, path string, mode uint32) error {
	// the host's users mean nothing in an image, and a directory with
	// setgid hands on its own group
	if err := unix.Fchown(fd, 0, 0); err != nil {
		return &os.PathError{Op: "chown", Path: path, Err: err}
	}

	// after the owner, whose change clears setuid, and the mode given at
	// create passes the umask first
	if err := unix.Fchmod(fd, mode); err != nil {
		return &os.PathError{Op: "chmod", Path: path, Err: err}
	}

	// the digest leaves times out, so a layer must not take them from the
	// moment it was built
	if err := unix.Futimes(fd, make([]unix.Timeval, 2)); err != nil {
		return &os.PathError{Op: "utimes", Path: path, Err: err}
	}

	return nil
}
