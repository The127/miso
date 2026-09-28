package place

import (
	"os"

	"golang.org/x/sys/unix"
)

// Node puts a device, fifo or socket of a kind, as a mode's type bits say,
// at a path of the image. It is open to its owner alone until it is given
// the mode it is copied with.
func (r *Root) Node(path string, kind uint32, device uint64) error {
	parent, name, err := r.parent(path)
	if err != nil {
		return err
	}

	defer func() { _ = unix.Close(parent) }()

	if err := vacate(parent, name, path); err != nil {
		return err
	}

	if err := unix.Mknodat(parent, name, kind&unix.S_IFMT|0o600, int(device)); err != nil { //nolint:gosec // the kernel keeps device numbers in 32 bits
		return &os.PathError{Op: "mknod", Path: path, Err: err}
	}

	return nil
}
