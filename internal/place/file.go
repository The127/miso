package place

import (
	"io"
	"os"

	"golang.org/x/sys/unix"
)

// File puts a file with a content at a path of the image.
func (r *Root) File(path string, mode uint32, content io.Reader) error {
	parent, name, err := r.parent(path)
	if err != nil {
		return err
	}

	if name == "" {
		return &os.PathError{Op: "create", Path: path, Err: unix.EISDIR}
	}

	defer func() { _ = unix.Close(parent) }()

	if err := vacate(parent, name, path); err != nil {
		return err
	}

	fd, err := unix.Openat(parent, name, unix.O_WRONLY|unix.O_CREAT|unix.O_EXCL|unix.O_NOFOLLOW|unix.O_CLOEXEC, mode)
	if err != nil {
		return &os.PathError{Op: "create", Path: path, Err: err}
	}

	file := os.NewFile(uintptr(fd), path)
	if _, err := io.Copy(file, content); err != nil {
		_ = file.Close()

		return err
	}

	if err := settle(fd, path, mode); err != nil {
		_ = file.Close()

		return err
	}

	return file.Close()
}
