package baseimage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Fetch downloads the image a name stands for and answers its digest, that
// of the bytes as they arrived.
func (c *Cache) Fetch(ctx context.Context, name string) (string, error) {
	url, known := c.sources[name]
	if !known {
		return "", fmt.Errorf("%s: %w", name, ErrUnknownBase)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	response, err := c.client.Do(request)
	if err != nil {
		return "", err
	}

	defer func() { _ = response.Body.Close() }()

	digest, err := c.store(response.Body)
	if err != nil {
		return "", err
	}

	return digest, c.remember(name, digest)
}

// store writes the bytes to a file that gets its name from their digest.
// Until the last byte is in, the file has a temporary name, so a file
// named by a digest is always whole.
func (c *Cache) store(body io.Reader) (string, error) {
	if err := os.MkdirAll(c.blobs(), 0o750); err != nil {
		return "", err
	}

	file, err := os.CreateTemp(c.blobs(), "download-*")
	if err != nil {
		return "", err
	}

	hash := sha256.New()
	if _, err := io.Copy(io.MultiWriter(file, hash), body); err != nil {
		_ = file.Close()
		_ = os.Remove(file.Name())

		return "", err
	}

	if err := file.Close(); err != nil {
		return "", err
	}

	digest := "sha256:" + hex.EncodeToString(hash.Sum(nil))

	return digest, os.Rename(file.Name(), c.blob(digest))
}

func (c *Cache) remember(name string, digest string) error {
	if err := os.MkdirAll(c.names(), 0o750); err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(c.names(), name), []byte(digest), 0o600)
}

func (c *Cache) blobs() string {
	return filepath.Join(c.dir, "sha256")
}

// blob is where the bytes with a digest live, "sha256:<hex>" as sha256/<hex>.
func (c *Cache) blob(digest string) string {
	return filepath.Join(c.dir, filepath.FromSlash(strings.Replace(digest, ":", "/", 1)))
}

func (c *Cache) names() string {
	return filepath.Join(c.dir, "names")
}
