package baseimage

import (
	"os"
	"path/filepath"
	"strings"
)

// Format is that of the image with a digest, as its source declared it.
func (c *Cache) Format(digest string) (string, error) {
	format, err := os.ReadFile(c.formatOf(digest))
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
