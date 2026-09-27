package baseimage

import (
	"errors"
	"io"
	"os"
)

// ErrFormat is an image that is not in the format its source declares.
var ErrFormat = errors.New("image is not in the format its source declares")

const qcow2Magic = "QFI\xfb"

// check fails for an image with a digest that does not start the way the
// format declared for it does.
func (c *Cache) check(digest, format string) error {
	head, err := start(c.blobs.Path(digest), len(qcow2Magic))
	if err != nil {
		return err
	}

	if !fits(head, format) {
		return ErrFormat
	}

	return nil
}

// fits tells whether an image starting with some bytes can be in a format.
func fits(start, format string) bool {
	switch format {
	case "qcow2":
		return start == qcow2Magic
	case "raw":
		// a raw image starting like qcow2 is a qcow2 image declared wrong
		return start != "" && start != qcow2Magic
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
