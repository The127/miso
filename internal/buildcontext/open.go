package buildcontext

import (
	"fmt"
	"io/fs"
	"os"
	"syscall"
)

// open opens a file of the build context to read its content. What the
// name is may have changed since it was looked at, so the open file is asked
// again. Opening a pipe would wait for a writer forever, unless it does not
// block.
func (d *Dir) open(name string, looked fs.FileInfo) (*os.File, error) {
	file, err := d.root.OpenFile(name, os.O_RDONLY|syscall.O_NONBLOCK, 0)
	if err != nil {
		return nil, err
	}

	info, err := file.Stat()
	if err != nil {
		_ = file.Close()

		return nil, err
	}

	if !info.Mode().IsRegular() {
		_ = file.Close()

		return nil, fmt.Errorf("%s: %w", name, ErrSpecialFile)
	}

	// the root follows a link the name became, as long as it stays inside,
	// and the thing it reaches is not the one looked at
	if !os.SameFile(looked, info) {
		_ = file.Close()

		return nil, fmt.Errorf("%s: %w", name, ErrSwapped)
	}

	return file, nil
}
