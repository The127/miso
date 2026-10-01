package layer

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/The127/miso/internal/lru"
)

func unfinished(name string) bool {
	return strings.HasPrefix(name, working) || strings.HasPrefix(name, scratch)
}

// Sweep removes what a builder stopped halfway left behind. Only while
// nothing uses the store, before the first layer is begun.
func (s *Store) Sweep() (lru.Swept, error) {
	var swept lru.Swept

	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return swept, err
	}

	for _, entry := range entries {
		if !unfinished(entry.Name()) {
			continue
		}

		path := filepath.Join(s.dir, entry.Name())

		size, err := usage(path)
		if err != nil {
			return swept, err
		}

		if err := os.RemoveAll(path); err != nil {
			return swept, err
		}

		swept.Count++
		swept.Bytes += size
	}

	return swept, nil
}
