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
	plain

	request build.Request
	check   Check
}

// needsCheck says whether the disk of a request is booted before it gets its
// name.
func needsCheck(request build.Request) bool {
	return len(request.Checks) > 0
}

// unchecked is the name a disk has until its checks passed.
func unchecked(name string) string {
	return path.Join(path.Dir(name), "."+path.Base(name)+".unchecked")
}

func (c checked) Close() error {
	if err := c.File.Close(); err != nil {
		return err
	}

	return inRoot(c.dir, func(root *os.Root) error {
		if err := c.check(filepath.Join(c.dir, c.name), c.request); err != nil {
			return errors.Join(err, root.Remove(c.name))
		}

		return root.Rename(c.name, c.request.Output)
	})
}

// OnlyChecked are outputs for the disks with checks alone and the files that
// checks of a later output boot with, so a build that writes no outputs still
// boots what it must check.
func OnlyChecked(outputs Outputs) Outputs {
	return func(request build.Request) (Output, error) {
		if !needsCheck(request) && !request.Needed {
			return nil, nil
		}

		return outputs(request)
	}
}
