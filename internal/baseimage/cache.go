package baseimage

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/The127/miso/internal/download"
)

// ErrUnknownBase is a name no source is known for. It does not say which,
// the FROM line a plan wraps it in already does.
var ErrUnknownBase = errors.New("unknown base image")

// Cache holds fetched images under one directory on the host.
type Cache struct {
	dir     string
	blobs   *download.Store
	sources map[string]string
}

// Open takes the directory the images live in and where each name comes
// from. The images themselves live in the store. It touches nothing yet.
func Open(dir string, blobs *download.Store, sources map[string]string) *Cache {
	return &Cache{dir: dir, blobs: blobs, sources: sources}
}

// Digest is that of the image a name stands for, or empty for one that is
// known but not fetched yet.
func (c *Cache) Digest(name string) (string, error) {
	if _, err := c.source(name); err != nil {
		return "", err
	}

	digest, err := os.ReadFile(filepath.Join(c.names(), name))
	if errors.Is(err, fs.ErrNotExist) {
		return "", nil
	}

	if err != nil {
		return "", err
	}

	// a name is only worth what it points at, and a build fetches again
	// what somebody cleaned away
	//nolint:gosec // the digest is what remember wrote, nothing a user names
	if _, err := os.Stat(c.blobs.Path(string(digest))); errors.Is(err, fs.ErrNotExist) {
		return "", nil
	}

	return string(digest), nil
}

// source is where a name comes from.
func (c *Cache) source(name string) (string, error) {
	url, known := c.sources[name]
	if !known {
		return "", ErrUnknownBase
	}

	return url, nil
}
