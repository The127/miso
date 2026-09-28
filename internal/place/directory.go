package place

import (
	"errors"
	"os"

	"golang.org/x/sys/unix"
)

// Directory puts a directory with a mode at a path of the image.
func (r *Root) Directory(path string, mode uint32) error {
	parent, name, err := r.parent(path)
	if err != nil {
		return err
	}

	// the root is always there, and keeps what it has as any other would
	if name == "" {
		return nil
	}

	defer func() { _ = unix.Close(parent) }()

	fd, err := makeDirectory(parent, name, path, mode)
	// one that is there keeps its mode, owner and time, so that COPY rootfs/
	// / does not hand the host's modes to /, /etc and the rest
	if errors.Is(err, unix.EEXIST) {
		fd, err = r.at(path)
	}

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
