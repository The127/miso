package builder

import (
	"errors"
	"os"
	"path"
	"path/filepath"

	"github.com/The127/miso/internal/build"
)

// Check boots a fetched disk and runs the checks of its request in it.
type Check func(disk string, request build.Request) error

// checked is the file of a disk with checks, under a name of its own until
// they passed, so a disk that never booted never shows under its output's
// name.
type checked struct {
	*os.File

	dir     string
	name    string
	request build.Request
	check   Check
}

// unchecked is the name a disk has until its checks passed.
func unchecked(name string) string {
	return path.Join(path.Dir(name), "."+path.Base(name)+".unchecked")
}

func (c checked) Close() error {
	if err := c.File.Close(); err != nil {
		return err
	}

	root, err := os.OpenRoot(c.dir)
	if err != nil {
		return err
	}

	defer func() { _ = root.Close() }()

	if err := c.check(filepath.Join(c.dir, c.name), c.request); err != nil {
		return errors.Join(err, root.Remove(c.name))
	}

	return root.Rename(c.name, c.request.Output)
}

// Discard throws the disk away without booting it.
func (c checked) Discard() error {
	if err := c.File.Close(); err != nil {
		return err
	}

	root, err := os.OpenRoot(c.dir)
	if err != nil {
		return err
	}

	defer func() { _ = root.Close() }()

	return root.Remove(c.name)
}

// OnlyChecked are outputs for the disks with checks alone and the files that
// checks of a later output boot with, so a build that writes no outputs still
// boots what it must check.
func OnlyChecked(outputs Outputs) Outputs {
	return func(request build.Request) (Output, error) {
		if len(request.Checks) == 0 && !request.Needed {
			return nil, nil
		}

		return outputs(request)
	}
}
