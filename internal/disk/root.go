package disk

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math"
)

const sector = 512

// ErrBrokenTable is a partition table no real disk can have.
var ErrBrokenTable = errors.New("broken partition table")

// rootX86_64 is the partition type of an x86-64 root file system in the
// order its bytes are on disk.
var rootX86_64 = []byte{
	0xe3, 0xbc, 0x68, 0x4f,
	0xcd, 0xe8,
	0xb1, 0x4d,
	0x96, 0xe7, 0xfb, 0xca, 0xf9, 0x84, 0xb7, 0x09,
}

// Partition is where a partition lies on its disk, in bytes.
type Partition struct {
	Offset int64
	Size   int64
}

// Root finds the root partition in the GPT of a disk.
func Root(r io.ReaderAt) (Partition, error) {
	header := make([]byte, 92)
	if _, err := r.ReadAt(header, sector); err != nil {
		return Partition{}, err
	}

	entries, err := lba(header[72:])
	if err != nil {
		return Partition{}, err
	}

	count := binary.LittleEndian.Uint32(header[80:])
	size := binary.LittleEndian.Uint32(header[84:])

	entry := make([]byte, size)
	for i := range int64(count) {
		if _, err := r.ReadAt(entry, entries*sector+i*int64(size)); err != nil {
			return Partition{}, err
		}

		if bytes.Equal(entry[:16], rootX86_64) {
			first, err := lba(entry[32:])
			if err != nil {
				return Partition{}, err
			}

			last, err := lba(entry[40:])
			if err != nil {
				return Partition{}, err
			}

			return Partition{Offset: first * sector, Size: (last - first + 1) * sector}, nil
		}
	}

	return Partition{}, nil
}

// lba is a sector number that still fits a byte offset.
func lba(b []byte) (int64, error) {
	n := binary.LittleEndian.Uint64(b)
	if n > math.MaxInt64/sector {
		return 0, ErrBrokenTable
	}

	return int64(n), nil
}
