package baseimage

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"
)

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

func (c *Cache) blobs() string {
	return filepath.Join(c.dir, "sha256")
}

// blob is where the bytes with a digest live, sha256:<hex> under sha256/.
func (c *Cache) blob(digest string) string {
	return filepath.Join(c.blobs(), strings.TrimPrefix(digest, "sha256:"))
}
