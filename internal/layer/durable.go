package layer

import (
	"errors"
	"os"

	"golang.org/x/sys/unix"
)

// syncFileSystem has the kernel put everything written to the file system
// of a directory on the disk.
func syncFileSystem(path string) error {
	dir, err := os.Open(path)
	if err != nil {
		return err
	}

	return errors.Join(unix.Syncfs(int(dir.Fd())), dir.Close())
}
