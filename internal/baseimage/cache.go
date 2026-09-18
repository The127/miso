package baseimage

import (
	"errors"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
)

// ErrUnknownBase is a name no source is known for. It does not say which,
// the FROM line a plan wraps it in already does.
var ErrUnknownBase = errors.New("unknown base image")

// Cache holds fetched images under one directory on the host.
type Cache struct {
	dir     string
	client  *http.Client
	sources map[string]string
}

// Open takes the directory the images live in, the client that fetches
// them, and where each name comes from. It touches nothing yet.
func Open(dir string, client *http.Client, sources map[string]string) *Cache {
	return &Cache{dir: dir, client: client, sources: sources}
}

// Digest is that of the image a name stands for, or empty for one that is
// known but not fetched yet.
func (c *Cache) Digest(name string) (string, error) {
	if _, known := c.sources[name]; !known {
		return "", ErrUnknownBase
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
	if _, err := os.Stat(c.blob(string(digest))); errors.Is(err, fs.ErrNotExist) {
		return "", nil
	}

	return string(digest), nil
}
