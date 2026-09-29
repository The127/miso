package layer

import (
	"io/fs"
	"path/filepath"

	"golang.org/x/sys/unix"
)

// usage is the bytes a directory takes on the disk. It counts blocks, not
// lengths, because layers hold sparse files.
func usage(dir string) (int64, error) {
	var total int64

	err := filepath.WalkDir(dir, func(path string, _ fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		var stat unix.Stat_t
		if err := unix.Lstat(path, &stat); err != nil {
			return err
		}

		total += stat.Blocks * 512

		return nil
	})

	return total, err
}
