package place

import (
	"errors"
	"io"
	"os"

	"golang.org/x/sys/unix"
)

// Sparse puts a copy of a regular file at a path of the image, range of
// data by range of data, so that its holes stay holes. A plain copy writes
// a hole as zeros on file systems like tmpfs, and a sparse image of a disk
// would then take all of its size.
func (r *Root) Sparse(path string, mode uint32, source *os.File) error {
	return r.create(path, mode, func(file *os.File) error {
		info, err := source.Stat()
		if err != nil {
			return err
		}

		fd := int(source.Fd())
		for at := int64(0); at < info.Size(); {
			data, err := unix.Seek(fd, at, unix.SEEK_DATA)
			// no data past at, the rest is a hole
			if errors.Is(err, unix.ENXIO) {
				break
			}

			if err != nil {
				return err
			}

			hole, err := unix.Seek(fd, data, unix.SEEK_HOLE)
			if err != nil {
				return err
			}

			if err := copyRange(file, source, data, hole-data); err != nil {
				return err
			}

			at = hole
		}

		return file.Truncate(info.Size())
	})
}

// copyRange copies a range of the source to the same place in the file,
// letting the kernel copy it without reading it into miso.
func copyRange(file, source *os.File, at, size int64) error {
	if _, err := source.Seek(at, io.SeekStart); err != nil {
		return err
	}

	if _, err := file.Seek(at, io.SeekStart); err != nil {
		return err
	}

	_, err := io.CopyN(file, source, size)

	return err
}
