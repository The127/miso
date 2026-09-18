package disk

import (
	"encoding/binary"
	"errors"
	"io"
)

// ErrUnknownFileSystem is a partition with no file system miso can mount.
var ErrUnknownFileSystem = errors.New("unknown file system")

// FileSystem names the file system on a partition the way mount(2) takes it.
func FileSystem(r io.ReaderAt) (string, error) {
	magic := make([]byte, 2)
	if _, err := r.ReadAt(magic, 1080); err != nil {
		return "", err
	}

	// ext2 and ext3 share the magic, and the ext4 driver mounts them too
	if binary.LittleEndian.Uint16(magic) == 0xef53 {
		return "ext4", nil
	}

	btrfs := make([]byte, 8)
	if _, err := r.ReadAt(btrfs, 65600); err != nil {
		return "", err
	}

	if string(btrfs) == "_BHRfS_M" {
		return "btrfs", nil
	}

	return "", ErrUnknownFileSystem
}
