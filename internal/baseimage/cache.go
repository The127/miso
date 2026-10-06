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
	sources map[string]Source
}

// Source is where a name comes from and the disk format of what is there,
// as QEMU names it.
type Source struct {
	URL    string
	Format string
}

// Open takes the directory the images live in and where each name comes
// from. The images themselves live in the store. It touches nothing yet.
func Open(dir string, blobs *download.Store, sources map[string]Source) *Cache {
	return &Cache{dir: dir, blobs: blobs, sources: sources}
}

// OpenFor opens the images miso knows for an architecture. amd64 keeps the
// directory it always had, so a cache from before other architectures is
// still found, and each other one keeps its names and formats in a directory
// of its own.
func OpenFor(dir, arch string, blobs *download.Store) *Cache {
	if arch != "amd64" {
		dir = filepath.Join(dir, arch)
	}

	return Open(dir, blobs, Known[arch])
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
func (c *Cache) source(name string) (Source, error) {
	source, known := c.sources[name]
	if !known {
		return Source{}, ErrUnknownBase
	}

	return source, nil
}
