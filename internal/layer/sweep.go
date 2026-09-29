package layer

import (
	"os"
	"path/filepath"
	"strings"
)

func unfinished(name string) bool {
	return strings.HasPrefix(name, working) || strings.HasPrefix(name, scratch)
}

// Swept is what a sweep removed.
type Swept struct {
	Count int
	Bytes int64
}

// Sweep removes what a builder stopped halfway left behind. Only while
// nothing uses the store, before the first layer is begun.
func (s *Store) Sweep() (Swept, error) {
	var swept Swept

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
