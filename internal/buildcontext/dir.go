package buildcontext

import (
	"errors"
	"os"
)

// ErrOutsideContext is a source that is not below the build context.
var ErrOutsideContext = errors.New("outside the build context")

// ErrThroughLink is a source whose path goes through a link. A link of the
// build context is never followed on the host.
var ErrThroughLink = errors.New("goes through a link")

// ErrSpecialFile is something in the build context that is no file, no
// directory and no link, such as a pipe or a device. A copy cannot carry it.
var ErrSpecialFile = errors.New("special file")

// Dir is a build context, a directory on the host.
type Dir struct {
	// the root keeps every path below the directory, also one that tries
	// to leave it through .. or a symlink
	root *os.Root
}

// Open takes the directory a build file sits in as its build context.
func Open(dir string) (*Dir, error) {
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, err
	}

	return &Dir{root: root}, nil
}

// Close lets go of the directory.
func (d *Dir) Close() error {
	return d.root.Close()
}
