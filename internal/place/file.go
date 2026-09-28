package place

import (
	"errors"
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

	// a file replaces what was there, so it takes neither a link's target
	// nor an old file's mode or other names
	name := filepath.Base(path)
	if err := unix.Unlinkat(parent, name, 0); err != nil && !errors.Is(err, unix.ENOENT) {
		return &os.PathError{Op: "remove", Path: path, Err: err}
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
