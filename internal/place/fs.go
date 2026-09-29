package place

import (
	"io/fs"
	"os"

	"golang.org/x/sys/unix"
)

// FS reads the image, with its links meaning places in the image.
func (r *Root) FS() fs.FS {
	return image{root: r}
}

type image struct {
	root *Root
}

func (i image) Open(name string) (fs.File, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrInvalid}
	}

	root, err := i.root.openRoot()
	if err != nil {
		return nil, err
	}

	defer func() { _ = unix.Close(root) }()

	// a FIFO of the image would wait for a writer that never comes
	fd, err := inImage(root, name, unix.O_RDONLY|unix.O_NONBLOCK|unix.O_CLOEXEC)
	if err != nil {
		return nil, &fs.PathError{Op: "open", Path: name, Err: err}
	}

	return os.NewFile(uintptr(fd), name), nil
}
