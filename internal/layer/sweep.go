package layer

import (
	"os"
	"path/filepath"
	"strings"
)

func unfinished(name string) bool {
	return strings.HasPrefix(name, working) || strings.HasPrefix(name, scratch)
}

// Sweep removes what a builder stopped halfway left behind. Only while
// nothing uses the store, before the first layer is begun.
func (s *Store) Sweep() error {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !unfinished(entry.Name()) {
			continue
		}

		if err := os.RemoveAll(filepath.Join(s.dir, entry.Name())); err != nil {
			return err
		}
	}

	return nil
}
