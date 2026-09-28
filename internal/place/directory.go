package place

import (
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

// Directory puts a directory with a mode at a path of the image.
func (r *Root) Directory(path string, mode uint32) error {
	parent, err := r.parent(path)
	if err != nil {
		return err
	}

	defer func() { _ = unix.Close(parent) }()

	fd, err := makeDirectory(parent, filepath.Base(path), path, mode)
	if err != nil {
		return err
	}

	return unix.Close(fd)
}

// makeDirectory makes a directory of the image and opens it.
func makeDirectory(parent int, name string, path string, mode uint32) (int, error) {
	if err := unix.Mkdirat(parent, name, 0o700); err != nil {
		return -1, &os.PathError{Op: "mkdir", Path: path, Err: err}
	}

	fd, err := unix.Openat(parent, name, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
	if err != nil {
		return -1, &os.PathError{Op: "open", Path: path, Err: err}
	}

	if err := settle(fd, path, mode); err != nil {
		_ = unix.Close(fd)

		return -1, err
	}

	return fd, nil
}
