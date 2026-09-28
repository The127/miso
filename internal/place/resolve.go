package place

import (
	"errors"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

// parent opens the directory a path of the image lives in, and makes it
// when it is missing.
func (r *Root) parent(path string) (int, error) {
	root, err := unix.Open(r.dir, unix.O_PATH|unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
	if err != nil {
		return -1, &os.PathError{Op: "open", Path: r.dir, Err: err}
	}

	defer func() { _ = unix.Close(root) }()

	return directory(root, filepath.Dir(path))
}

// directory opens a directory of the image, made open to all and root's
// with its missing parents when it is not there. An image's links mean
// places in the image, an absolute one too, so each is resolved as if the
// image's root were the root of all. Followed as they are, /var/run -> /run
// in Debian would reach the builder VM's own /run.
func directory(root int, dir string) (int, error) {
	fd, err := unix.Openat2(root, dir, &unix.OpenHow{
		Flags:   unix.O_RDONLY | unix.O_DIRECTORY | unix.O_CLOEXEC,
		Resolve: unix.RESOLVE_IN_ROOT,
	})
	if err == nil {
		return fd, nil
	}

	if !errors.Is(err, unix.ENOENT) || dir == "/" {
		return -1, &os.PathError{Op: "open", Path: dir, Err: err}
	}

	parent, err := directory(root, filepath.Dir(dir))
	if err != nil {
		return -1, err
	}

	defer func() { _ = unix.Close(parent) }()

	return makeDirectory(parent, filepath.Base(dir), dir, 0o755)
}
