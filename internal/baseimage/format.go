package baseimage

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// ErrNotFetched is a digest of no image this cache fetched.
var ErrNotFetched = errors.New("no image fetched with that digest")

// ErrFormat is an image that is not in the format its source declares.
var ErrFormat = errors.New("image is not in the format its source declares")

// Format is that of the image with a digest, as its source declared it.
func (c *Cache) Format(digest string) (string, error) {
	format, err := os.ReadFile(c.formatOf(digest))
	if errors.Is(err, fs.ErrNotExist) {
		return "", ErrNotFetched
	}

	if err != nil {
		return "", err
	}

	return string(format), nil
}

func (c *Cache) record(digest, format string) error {
	if err := os.MkdirAll(filepath.Dir(c.formatOf(digest)), 0o750); err != nil {
		return err
	}

	return os.WriteFile(c.formatOf(digest), []byte(format), 0o600)
}

// formatOf is the file holding the format of the image with a digest.
func (c *Cache) formatOf(digest string) string {
	return filepath.Join(c.dir, "formats", strings.TrimPrefix(digest, "sha256:"))
}

// check fails for an image with a digest that does not start the way the
// format declared for it does.
func (c *Cache) check(digest, format string) error {
	if format != "raw" && format != "qcow2" {
		return ErrFormat
	}

	image, err := os.Open(c.blobs.Path(digest))
	if err != nil {
		return err
	}

	defer func() { _ = image.Close() }()

	// an image shorter than the magic starts with none
	magic := make([]byte, 4)
	read, err := io.ReadFull(image, magic)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
		return err
	}

	// a raw image starting like qcow2 is a qcow2 image declared wrong
	if (string(magic[:read]) == "QFI\xfb") != (format == "qcow2") {
		return ErrFormat
	}

	return nil
}
