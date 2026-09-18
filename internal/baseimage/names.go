package baseimage

import (
	"os"
	"path/filepath"
)

func (c *Cache) remember(name string, digest string) error {
	if err := os.MkdirAll(c.names(), 0o750); err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(c.names(), name), []byte(digest), 0o600)
}

func (c *Cache) names() string {
	return filepath.Join(c.dir, "names")
}
