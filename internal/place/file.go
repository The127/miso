package place

import (
	"io"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

// File puts a file with a content at a path of the image.
func (r *Root) File(path string, mode uint32, content io.Reader) error {
	parent, err := r.parent(path)
	if err != nil {
		return err
	}

	defer func() { _ = unix.Close(parent) }()

	fd, err := unix.Openat(parent, filepath.Base(path), unix.O_WRONLY|unix.O_CREAT|unix.O_TRUNC|unix.O_NOFOLLOW|unix.O_CLOEXEC, mode)
	if err != nil {
		return &os.PathError{Op: "create", Path: path, Err: err}
	}

	file := os.NewFile(uintptr(fd), path)
	if _, err := io.Copy(file, content); err != nil {
		_ = file.Close()

		return err
	}

	// the host's users mean nothing in an image, and a directory with
	// setgid hands on its own group
	if err := unix.Fchown(fd, 0, 0); err != nil {
		_ = file.Close()

		return &os.PathError{Op: "chown", Path: path, Err: err}
	}

	// after the owner, whose change clears setuid, and the mode given at
	// create passes the umask first
	if err := unix.Fchmod(fd, mode); err != nil {
		_ = file.Close()

		return &os.PathError{Op: "chmod", Path: path, Err: err}
	}

	return file.Close()
}
