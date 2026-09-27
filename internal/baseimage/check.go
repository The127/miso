package baseimage

import (
	"encoding/binary"
	"errors"
	"io"
	"os"
	"strings"
)

// ErrFormat is an image that is not in the format its source declares.
var ErrFormat = errors.New("image is not in the format its source declares")

// ErrExternalFile is a qcow2 image that names a file outside itself.
var ErrExternalFile = errors.New("qcow2 image names a file outside itself")

const qcow2Magic = "QFI\xfb"

// qcow2Header is as much of a qcow2 header as a check reads, through its
// incompatible features.
const qcow2Header = 80

// qcow2DataFile is the incompatible feature of a qcow2 image whose data
// lives in a file of its own.
const qcow2DataFile = 1 << 2

// check fails for an image with a digest that does not start the way the
// format declared for it does.
func (c *Cache) check(digest, format string) error {
	head, err := start(c.blobs.Path(digest), qcow2Header)
	if err != nil {
		return err
	}

	if !fits(head, format) {
		return ErrFormat
	}

	// QEMU would open the host's file the downloaded image names
	if format == "qcow2" && binary.BigEndian.Uint64([]byte(head[8:16])) != 0 {
		return ErrExternalFile
	}

	// a version 2 header ends before the features, what follows is not one
	version := binary.BigEndian.Uint32([]byte(head[4:8]))
	if format == "qcow2" && version >= 3 && binary.BigEndian.Uint64([]byte(head[72:80]))&qcow2DataFile != 0 {
		return ErrExternalFile
	}

	return nil
}

// fits tells whether an image starting with some bytes can be in a format.
func fits(start, format string) bool {
	switch format {
	case "qcow2":
		return len(start) == qcow2Header && strings.HasPrefix(start, qcow2Magic)
	case "raw":
		// a raw image starting like qcow2 is a qcow2 image declared wrong
		return start != "" && !strings.HasPrefix(start, qcow2Magic)
	default:
		return false
	}
}

// start is up to the first n bytes of a file, fewer for a shorter one.
func start(path string, n int) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}

	defer func() { _ = file.Close() }()

	bytes := make([]byte, n)
	read, err := io.ReadFull(file, bytes)
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
		return "", err
	}

	return string(bytes[:read]), nil
}
