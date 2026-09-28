package place

import (
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

// parent opens the directory a path of the image lives in. An image's links
// mean places in the image, an absolute one too, so each is resolved as if
// the image's root were the root of all. Followed as they are, /var/run ->
// /run in Debian would reach the builder VM's own /run.
func (r *Root) parent(path string) (int, error) {
	root, err := unix.Open(r.dir, unix.O_PATH|unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
	if err != nil {
		return -1, &os.PathError{Op: "open", Path: r.dir, Err: err}
	}

	defer func() { _ = unix.Close(root) }()

	parent, err := unix.Openat2(root, filepath.Dir(path), &unix.OpenHow{
		Flags:   unix.O_PATH | unix.O_DIRECTORY | unix.O_CLOEXEC,
		Resolve: unix.RESOLVE_IN_ROOT,
	})
	if err != nil {
		return -1, &os.PathError{Op: "open", Path: filepath.Dir(path), Err: err}
	}

	return parent, nil
}
